package watcher

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog"

	"github.com/npavlov/go-metrics-service/internal/agent/config"
	"github.com/npavlov/go-metrics-service/internal/model"
)

// ErrPostRequestFailed Define a static error for failed POST requests.
var ErrPostRequestFailed = errors.New("failed to send post request")

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
			if mr.cfg.UseBatch {
				mr.SendMetricsBatch(ctx)
			} else {
				mr.SendMetrics(ctx)
			}
			time.Sleep(mr.cfg.ReportIntervalDur)
		}
	}
}

// SendMetrics sends the collected watcher to the server.
func (mr *MetricReporter) SendMetrics(ctx context.Context) {
	mr.mux.Lock()
	defer mr.mux.Unlock()

	for _, metric := range *mr.metrics {
		func() {
			if metric.Delta == nil && metric.Value == nil {
				return
			}

			url := mr.cfg.Address + "/update/"

			// Setting up context with a timeout
			timeoutCtx, cancel := context.WithTimeout(ctx, 1*time.Second)
			defer cancel()

			response, err := mr.sendPostRequest(timeoutCtx, url, metric)
			if err != nil {
				mr.l.Info().Err(err).Msg("Error sending metric")
			}
			mr.read(response)
		}()
	}
}

// SendMetricsBatch sends the collected watcher to the server.
func (mr *MetricReporter) SendMetricsBatch(ctx context.Context) {
	mr.mux.Lock()
	defer mr.mux.Unlock()

	url := mr.cfg.Address + "/updates/"

	// Setting up context with a timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	if len(*mr.metrics) == 0 {
		return
	}

	response, err := mr.sendPostRequest(timeoutCtx, url, *mr.metrics)
	if err != nil {
		mr.l.Info().Err(err).Msg("Error sending metrics batch")
	}
	mr.readMany(response)
}

func (mr *MetricReporter) sendPostRequest(ctx context.Context, url string, data interface{}) ([]byte, error) {
	// Marshal the metric to JSON
	payload, err := json.Marshal(data)
	if err != nil {
		mr.l.Error().Err(err).Msg("Failed to marshal metric")

		return nil, err
	}

	compressed, err := mr.compressRequest(payload)
	if err != nil {
		mr.l.Error().Err(err).Msg("Failed to compress payload")

		return nil, err
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

		return nil, err
	}

	if resp.StatusCode() != http.StatusOK {
		mr.l.Error().Int("statusCode", resp.StatusCode()).Msg("Failed to send post request")

		return nil, ErrPostRequestFailed
	}

	responseBody, err := mr.decompressResult(resp.RawBody())
	if err != nil {
		mr.l.Error().Err(err).Msg("Failed to decompress response")

		return nil, err
	}

	return responseBody, nil
}

func (mr *MetricReporter) read(data []byte) {
	// Unmarshal the decompressed response into a Metric struct
	var rMetric model.Metric
	err := json.Unmarshal(data, &rMetric)
	if err != nil {
		mr.l.Error().Err(err).Msg("Failed to unmarshal metric")

		return
	}

	// Log the successful transmission
	mr.l.Info().
		Interface("metric", rMetric).
		Msg("Server response")
}

func (mr *MetricReporter) readMany(data []byte) {
	// Unmarshal the decompressed response into a Metric struct
	var rMetrics []model.Metric
	err := json.Unmarshal(data, &rMetrics)
	if err != nil {
		mr.l.Error().Err(err).Msg("Failed to unmarshal metric")

		return
	}

	// Log the successful transmission
	mr.l.Info().
		Interface("metrics", rMetrics).
		Msg("Server response")
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
