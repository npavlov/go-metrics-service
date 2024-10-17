package main

import (
	"context"
	"fmt"
	"os"
	"runtime/debug"
	"sync"

	"github.com/joho/godotenv"
	"github.com/npavlov/go-metrics-service/internal/agent/config"
	"github.com/npavlov/go-metrics-service/internal/agent/stats"
	"github.com/npavlov/go-metrics-service/internal/agent/watcher"
	"github.com/npavlov/go-metrics-service/internal/logger"
	"github.com/npavlov/go-metrics-service/internal/utils"
	"github.com/rs/zerolog"
)

func main() {
	l := logger.NewLogger().SetLogLevel(zerolog.DebugLevel).Get()

	defer func() {
		// Recover from panic if one occurred. Log the error and exit.
		if r := recover(); r != nil {
			l.Fatal().
				Str("error", fmt.Sprintf("%v", r)).
				Bytes("stack", debug.Stack()).
				Msg("Fatal error encountered")
			os.Exit(1)
		}
	}()

	err := godotenv.Load("agent.env")
	if err != nil {
		l.Fatal().Msg("Error loading agent.env file")
	}

	cfg := config.NewConfigBuilder().
		FromEnv().
		FromFlags().Build()

	metrics := stats.NewStats().StatsToMetrics()
	mux := sync.RWMutex{}
	// WaitGroup to wait for all goroutines to complete
	var wg sync.WaitGroup

	ctx := utils.WithSignalCancel(context.Background())

	l.Info().
		Str("server_address", cfg.Address).
		Msg("Endpoint address set")
	var collector watcher.Collector = watcher.NewMetricCollector(&metrics, &mux, cfg)
	var reporter watcher.Reporter = watcher.NewMetricReporter(&metrics, &mux, cfg)

	l.Info().
		Int64("polling_time", cfg.PollInterval).
		Int64("reporting_time", cfg.ReportInterval).
		Msg("Polling and reporting times set")

	// Start watcher collection
	wg.Add(1)
	go collector.StartCollector(ctx, &wg)

	// Start watcher reporting
	wg.Add(1)
	go reporter.StartReporter(ctx, &wg)

	l.Info().Msg("Application started")
	utils.WaitForShutdown(&wg)
	l.Info().Msg("Application stopped gracefully")
}
