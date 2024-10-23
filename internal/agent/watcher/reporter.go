package watcher

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog"

	"github.com/npavlov/go-metrics-service/internal/agent/config"
	"github.com/npavlov/go-metrics-service/internal/model"
)

// Reporter interface defines the contract for sending watcher.
type Reporter interface {
	SendMetrics(ctx context.Context)
	StartReporter(ctx context.Context, wg *sync.WaitGroup)
}

// MetricReporter implements the Reporter interface.
type MetricReporter struct {
	metrics *[]model.Metric
	mux     *sync.RWMutex
	cfg     *config.Config
	l       *zerolog.Logger
}

func NewMetricReporter(m *[]model.Metric, mutex *sync.RWMutex, cfg *config.Config, l *zerolog.Logger) *MetricReporter {
	return &MetricReporter{
		metrics: m,
		mux:     mutex,
		cfg:     cfg,
		l:       l,
	}
}

func (mr *MetricReporter) StartReporter(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			mr.l.Info().Msg("Stopping watcher reporting")

			return
		default:
			// Add your watcher reporting logic here
			mr.SendMetrics(ctx)
			time.Sleep(time.Duration(mr.cfg.ReportInterval) * time.Second)
		}
	}
}

// SendMetrics sends the collected watcher to the server.
func (mr *MetricReporter) SendMetrics(ctx context.Context) {
	mr.mux.Lock()
	defer mr.mux.Unlock()

	for _, metric := range *mr.metrics {
		if metric.Delta == nil && metric.Value == nil {
			continue
		}

		url := mr.cfg.Address + "/update/"
		mr.sendPostRequest(ctx, url, metric)
	}
}

func (mr *MetricReporter) sendPostRequest(ctx context.Context, url string, metric model.Metric) {
	// Marshal the metric to JSON
	payload, err := json.Marshal(&metric)
	if err != nil {
		mr.l.Error().Err(err).Msg("Failed to marshal metric")

		return
	}

	compressed, err := mr.compressRequest(payload)
	if err != nil {
		mr.l.Error().Err(err).Msg("Failed to compress payload")

		return
	}

	// Create a new Resty client and send the POST request
	client := resty.New()
	resp, err := client.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetHeader("Content-Encoding", "gzip").
		SetHeader("Accept-Encoding", "gzip").
		SetBody(compressed.Bytes()).
		SetDoNotParseResponse(true).
		Post(url)
	if err != nil {
		mr.l.Error().Err(err).Msg("Failed to send post request")

		return
	}

	if resp.StatusCode() != http.StatusOK {
		mr.l.Error().Int("statusCode", resp.StatusCode()).Msg("Failed to send post request")

		return
	}

	responseBody, err := mr.decompressResult(resp.RawBody())
	if err != nil {
		mr.l.Error().Err(err).Msg("Failed to decompress response")
	}

	// Unmarshal the decompressed response into a Metric struct
	var rMetric model.Metric
	err = json.Unmarshal(responseBody, &rMetric)
	if err != nil {
		mr.l.Error().Err(err).Msg("Failed to unmarshal metric")

		return
	}

	// Log the successful transmission
	mr.l.Info().
		Str("url", url).
		Str("status", resp.Status()).
		Interface("metric", rMetric).
		Msg("Metric is sent")
}

func (mr *MetricReporter) compressRequest(data []byte) (*bytes.Buffer, error) {
	// Compress the payload using gzip
	var compressedPayload bytes.Buffer
	gzipWriter := gzip.NewWriter(&compressedPayload)
	_, err := gzipWriter.Write(data)
	if err != nil {
		return nil, err
	}
	err = gzipWriter.Close()
	if err != nil {
		return nil, err
	}

	return &compressedPayload, nil
}

func (mr *MetricReporter) decompressResult(body io.ReadCloser) ([]byte, error) {
	// Decompress the gzipped response
	reader, err := gzip.NewReader(body)
	if err != nil {
		return nil, err
	}
	defer func(reader *gzip.Reader) {
		_ = reader.Close()
	}(reader)

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	return data, nil
}
