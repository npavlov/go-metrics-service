package jsonsender

import (
	"context"
	"net/http"

	"github.com/go-resty/resty/v2"
	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	"github.com/npavlov/go-metrics-service/internal/agent/config"
	au "github.com/npavlov/go-metrics-service/internal/agent/utils"
	"github.com/npavlov/go-metrics-service/internal/server/db"
	"github.com/npavlov/go-metrics-service/internal/utils"
	"github.com/npavlov/go-metrics-service/pkg/crypto"
)

var ErrPostRequestFailed = errors.New("failed to send post request")

type JSONSender struct {
	cfg        *config.Config
	l          *zerolog.Logger
	json       jsoniter.API
	encryption *crypto.Encryption
	ip         string
}

func NewSender(cfg *config.Config, logger *zerolog.Logger) *JSONSender {
	sender := &JSONSender{
		cfg:        cfg,
		l:          logger,
		json:       jsoniter.ConfigCompatibleWithStandardLibrary,
		encryption: nil,
		ip:         au.GetLocalIP(logger),
	}

	if cfg.CryptoKey == "" {
		return sender
	}

	encryption, err := crypto.NewEncryption(cfg.CryptoKey)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to create encryption") // Consider Panic or returning an error
	}

	sender.encryption = encryption

	return sender
}

func (rh *JSONSender) Close() {
	// for model consistency
}

func (rh *JSONSender) SendMetric(ctx context.Context, metric db.Metric) (*db.Metric, error) {
	url := rh.cfg.Address + "/update/"

	data, err := rh.sendPostRequest(ctx, url, metric)
	if err != nil {
		return nil, err
	}

	return rh.read(data)
}

func (rh *JSONSender) SendMetricsBatch(ctx context.Context, metrics []db.Metric) ([]db.Metric, error) {
	url := rh.cfg.Address + "/updates/"

	data, err := rh.sendPostRequest(ctx, url, metrics)
	if err != nil {
		return nil, err
	}

	return rh.readMany(data)
}

func (rh *JSONSender) sendPostRequest(ctx context.Context, url string, data interface{}) ([]byte, error) {
	payload, err := rh.json.Marshal(data)
	if err != nil {
		rh.l.Error().Err(err).Msg("Failed to marshal metric")

		return nil, errors.Wrap(err, "failed to marshal metric")
	}

	compressed, err := au.Compress(payload)
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
		SetHeader("X-Real-IP", rh.ip).
		SetBody(compressed.Bytes()).
		SetDoNotParseResponse(true)

	if rh.cfg.Key != "" {
		hash := utils.CalculateHash(rh.cfg.Key, payload)
		request.SetHeader("HashSHA256", hash)
	}

	if rh.encryption != nil {
		encryptedPayload, err := rh.encryption.Encrypt(compressed.Bytes())
		if err != nil {
			return nil, errors.Wrap(err, "failed to encrypt metric")
		}
		request.SetBody(encryptedPayload)
		request.SetHeader("X-Encrypted", "true")
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
	responseBody, err := au.DecompressResult(resp.RawBody())
	if err != nil {
		rh.l.Error().Err(err).Msg("Failed to decompress response")

		return nil, errors.Wrapf(err, "Failed to decompress response")
	}

	return responseBody, nil
}

func (rh *JSONSender) read(data []byte) (*db.Metric, error) {
	// Unmarshal the decompressed response into a Metric struct
	var rMetric db.Metric
	err := rh.json.Unmarshal(data, &rMetric)
	if err != nil {
		rh.l.Error().Err(err).Msg("Failed to unmarshal metric")

		return nil, errors.Wrap(err, "failed to unmarshal metric")
	}

	return &rMetric, nil
}

func (rh *JSONSender) readMany(data []byte) ([]db.Metric, error) {
	// Unmarshal the decompressed response into a Metric struct
	var rMetrics []db.Metric
	err := rh.json.Unmarshal(data, &rMetrics)
	if err != nil {
		rh.l.Error().Err(err).Msg("Failed to unmarshal metric")

		return nil, errors.Wrapf(err, "Failed to unmarshal metric")
	}

	return rMetrics, nil
}
