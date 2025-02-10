package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/stretchr/testify/assert"

	testutils "github.com/npavlov/go-metrics-service/internal/test_utils"

	"github.com/npavlov/go-metrics-service/internal/agent/config"
	"github.com/npavlov/go-metrics-service/internal/agent/utils"
	"github.com/npavlov/go-metrics-service/internal/server/db"
)

func TestHandlePanic(t *testing.T) {
	t.Parallel()

	log := testutils.GetTLogger()
	defer func() {
		if r := recover(); r != nil {
			assert.NotNil(t, r)
		}
	}()
	handlePanic(log)
	panic("test panic")
}

func TestLoadConfig(t *testing.T) {
	t.Parallel()

	log := testutils.GetTLogger()

	cfg := loadConfig(log)
	assert.NotNil(t, cfg)
}

func TestRunAgent(t *testing.T) {
	t.Parallel()

	log := testutils.GetTLogger()
	json := jsoniter.ConfigCompatibleWithStandardLibrary
	cfg := &config.Config{
		PollIntervalDur:   100 * time.Millisecond,
		ReportIntervalDur: 100 * time.Millisecond,
		RateLimit:         1,
		UseBatch:          true,
	}

	// Mock server to simulate the Sender
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		var receivedMetrics []db.Metric
		result, _ := utils.DecompressResult(request.Body)
		err := json.Unmarshal(result, &receivedMetrics)
		assert.NoError(t, err)
		assert.NotEmpty(t, receivedMetrics)

		writer.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(writer).Encode(receivedMetrics)
	}))
	defer server.Close()

	cfg.Address = server.URL

	ctx, cancel := context.WithCancel(context.Background())

	go runAgent(ctx, cfg, log)

	// Allow some time for the reporter to process
	time.Sleep(300 * time.Millisecond)
	cancel()
}
