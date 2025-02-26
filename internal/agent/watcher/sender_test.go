package watcher_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	jsoniter "github.com/json-iterator/go"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/npavlov/go-metrics-service/pkg/crypto"

	"github.com/npavlov/go-metrics-service/internal/agent/utils"

	"github.com/npavlov/go-metrics-service/internal/agent/config"
	"github.com/npavlov/go-metrics-service/internal/agent/watcher"
	"github.com/npavlov/go-metrics-service/internal/server/db"
)

func TestSendMetric(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	logger := zerolog.Nop()
	cfg := &config.Config{
		Address: "http://localhost",
	}
	json := jsoniter.ConfigCompatibleWithStandardLibrary

	metric := db.NewMetric("test_metric", "gauge", nil, float64Ptr(3.14))

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assert.Equal(t, "/update/", request.URL.Path)
		assert.Equal(t, "gzip", request.Header.Get("Content-Encoding"))

		body, _ := utils.DecompressResult(request.Body)
		var receivedMetric db.Metric
		err := json.Unmarshal(body, &receivedMetric)
		assert.NoError(t, err)

		assert.Equal(t, metric, &receivedMetric)

		// Respond with the same metric
		payload, _ := json.Marshal(receivedMetric)
		responseData, _ := utils.Compress(payload)
		writer.Header().Set("Content-Encoding", "gzip")
		writer.WriteHeader(http.StatusOK)
		_, _ = writer.Write(responseData.Bytes())
	}))
	defer server.Close()

	cfg.Address = server.URL
	sender := watcher.NewSender(cfg, &logger)

	result, err := sender.SendMetric(ctx, *metric)
	require.NoError(t, err)
	assert.Equal(t, metric, result)
}

func TestSendMetricsBatch(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	logger := zerolog.Nop()
	cfg := &config.Config{
		Address: "http://localhost",
	}
	json := jsoniter.ConfigCompatibleWithStandardLibrary

	metrics := []db.Metric{
		*db.NewMetric("metric1", "counter", int64Ptr(1), nil),
		*db.NewMetric("metric2", "gauge", nil, float64Ptr(2.71)),
	}

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assert.Equal(t, "/updates/", request.URL.Path)
		assert.Equal(t, "gzip", request.Header.Get("Content-Encoding"))

		body, _ := utils.DecompressResult(request.Body)
		var receivedMetrics []db.Metric
		err := json.Unmarshal(body, &receivedMetrics)
		assert.NoError(t, err)

		assert.Equal(t, metrics, receivedMetrics)

		// Respond with the same metrics
		payload, _ := json.Marshal(receivedMetrics)
		responseData, _ := utils.Compress(payload)
		writer.Header().Set("Content-Encoding", "gzip")
		writer.WriteHeader(http.StatusOK)
		_, _ = writer.Write(responseData.Bytes())
	}))
	defer server.Close()

	cfg.Address = server.URL
	sender := watcher.NewSender(cfg, &logger)

	result, err := sender.SendMetricsBatch(ctx, metrics)
	require.NoError(t, err)
	assert.Equal(t, metrics, result)
}

func TestSendMetricError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	logger := zerolog.Nop()
	cfg := &config.Config{
		Address: "http://localhost",
	}

	metric := db.NewMetric("test_metric", "gauge", nil, float64Ptr(3.14))

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	cfg.Address = server.URL
	sender := watcher.NewSender(cfg, &logger)

	_, err := sender.SendMetric(ctx, *metric)
	assert.ErrorIs(t, err, watcher.ErrPostRequestFailed)
}

func TestSendMetricsBatchError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	logger := zerolog.Nop()
	cfg := &config.Config{
		Address: "http://localhost",
	}

	metrics := []db.Metric{
		*db.NewMetric("metric1", "counter", int64Ptr(1), nil),
		*db.NewMetric("metric2", "gauge", nil, float64Ptr(2.71)),
	}

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	cfg.Address = server.URL
	sender := watcher.NewSender(cfg, &logger)

	_, err := sender.SendMetricsBatch(ctx, metrics)
	assert.ErrorIs(t, err, watcher.ErrPostRequestFailed)
}

func TestSendEncryptedMetric(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	logger := zerolog.Nop()
	cfg := &config.Config{
		Address:   "http://localhost",
		CryptoKey: "testdata/test_public.key",
	}

	json := jsoniter.ConfigCompatibleWithStandardLibrary
	decryption, err := crypto.NewDecryption("testdata/test_private.key")
	require.NoError(t, err)

	metric := db.NewMetric("encrypted_metric", "gauge", nil, float64Ptr(9.81))

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assert.Equal(t, "/update/", request.URL.Path)
		assert.Equal(t, "true", request.Header.Get("X-Encrypted"))

		// Read and decrypt payload
		encryptedBody, _ := io.ReadAll(request.Body)
		decryptedBody, err := decryption.Decrypt(encryptedBody)
		assert.NoError(t, err)
		reader := io.NopCloser(bytes.NewBuffer(decryptedBody))
		body, err := utils.DecompressResult(reader)
		assert.NoError(t, err)

		var receivedMetric db.Metric
		err = json.Unmarshal(body, &receivedMetric)
		assert.NoError(t, err)

		assert.Equal(t, metric, &receivedMetric)

		// Respond with the same encrypted metric
		payload, _ := json.Marshal(receivedMetric)
		compressedResponse, _ := utils.Compress(payload)
		assert.NoError(t, err)

		writer.Header().Set("Content-Encoding", "gzip")
		writer.Header().Set("X-Encrypted", "true")
		writer.WriteHeader(http.StatusOK)
		_, _ = writer.Write(compressedResponse.Bytes())
	}))
	defer server.Close()

	cfg.Address = server.URL
	sender := watcher.NewSender(cfg, &logger)

	result, err := sender.SendMetric(ctx, *metric)
	require.NoError(t, err)
	assert.Equal(t, metric, result)
}

func float64Ptr(f float64) *float64 {
	return &f
}

func int64Ptr(i int64) *int64 {
	return &i
}
