package watcher

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"

	"github.com/go-resty/resty/v2"
	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	"github.com/npavlov/go-metrics-service/internal/agent/config"
	"github.com/npavlov/go-metrics-service/internal/agent/utils"
	"github.com/npavlov/go-metrics-service/internal/server/db"
)

var ErrPostRequestFailed = errors.New("failed to send post request")

type Result struct {
	Metric  *db.Metric  // Single metric (if applicable)
	Metrics []db.Metric // Array of stats (if applicable)
	Error   error       // Error (if any)
}

type Sender struct {
	cfg  *config.Config
	l    *zerolog.Logger
	json jsoniter.API
}

func NewSender(cfg *config.Config, logger *zerolog.Logger) *Sender {
	return &Sender{
		cfg:  cfg,
		l:    logger,
		json: jsoniter.ConfigCompatibleWithStandardLibrary,
	}
}

func (rh *Sender) SendMetric(ctx context.Context, metric db.Metric) (*db.Metric, error) {
	url := rh.cfg.Address + "/update/"

	data, err := rh.sendPostRequest(ctx, url, metric)
	if err != nil {
		return nil, err
	}

	return rh.read(data)
}

func (rh *Sender) SendMetricsBatch(ctx context.Context, metrics []db.Metric) ([]db.Metric, error) {
	url := rh.cfg.Address + "/updates/"

	data, err := rh.sendPostRequest(ctx, url, metrics)
	if err != nil {
		return nil, err
	}

	return rh.readMany(data)
}

func (rh *Sender) sendPostRequest(ctx context.Context, url string, data interface{}) ([]byte, error) {
	payload, err := rh.json.Marshal(data)
	if err != nil {
		rh.l.Error().Err(err).Msg("Failed to marshal metric")

		return nil, errors.Wrap(err, "failed to marshal metric")
	}

	compressed, err := utils.Compress(payload)
	if err != nil {
		rh.l.Error().Err(err).Msg("Failed to compress payload")

		return nil, errors.Wrap(err, "Failed to compress payload")
	}

	client := resty.New()
	request := client.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetHeader("Content-Encoding", "gzip").
		SetHeader("Accept-Encoding", "gzip").
		SetBody(compressed.Bytes()).
		SetDoNotParseResponse(true)

	if rh.cfg.Key != "" {
		hash := rh.calculateHash(payload)
		request.SetHeader("HashSHA256", hash)
	}

	resp, err := request.Post(url)
	if err != nil {
		rh.l.Error().Err(err).Msg("Failed to send post request")

		return nil, errors.Wrap(err, "Failed to send post request")
	}

	if resp.StatusCode() != http.StatusOK {
		rh.l.Error().Int("statusCode", resp.StatusCode()).Msg("Failed to send post request")

		return nil, ErrPostRequestFailed
	}

	// Handle the response based on whether it is a single metric or an array
	responseBody, err := utils.DecompressResult(resp.RawBody())
	if err != nil {
		rh.l.Error().Err(err).Msg("Failed to decompress response")

		return nil, errors.Wrapf(err, "Failed to decompress response")
	}

	return responseBody, nil
}

func (rh *Sender) calculateHash(payload []byte) string {
	h := hmac.New(sha256.New, []byte(rh.cfg.Key))
	h.Write(payload)

	return hex.EncodeToString(h.Sum(nil))
}

func (rh *Sender) read(data []byte) (*db.Metric, error) {
	// Unmarshal the decompressed response into a Metric struct
	var rMetric db.Metric
	err := rh.json.Unmarshal(data, &rMetric)
	if err != nil {
		rh.l.Error().Err(err).Msg("Failed to unmarshal metric")

		return nil, errors.Wrap(err, "failed to unmarshal metric")
	}

	return &rMetric, nil
}

func (rh *Sender) readMany(data []byte) ([]db.Metric, error) {
	// Unmarshal the decompressed response into a Metric struct
	var rMetrics []db.Metric
	err := rh.json.Unmarshal(data, &rMetrics)
	if err != nil {
		rh.l.Error().Err(err).Msg("Failed to unmarshal metric")

		return nil, errors.Wrapf(err, "Failed to unmarshal metric")
	}

	return rMetrics, nil
}
