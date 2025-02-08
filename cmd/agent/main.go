//nolint:gochecknoglobals
package main

import (
	"context"
	"fmt"
	"runtime/debug"
	"sync"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog"

	"github.com/npavlov/go-metrics-service/internal/agent/config"
	"github.com/npavlov/go-metrics-service/internal/agent/watcher"
	"github.com/npavlov/go-metrics-service/internal/domain"
	"github.com/npavlov/go-metrics-service/internal/logger"
	"github.com/npavlov/go-metrics-service/internal/server/db"
	"github.com/npavlov/go-metrics-service/internal/utils"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {
	log := logger.NewLogger(zerolog.DebugLevel).Get()

	log.Info().Str("buildVersion", buildVersion).
		Str("buildCommit", buildCommit).
		Str("buildDate", buildDate).Msg("Starting agent")

	defer func() {
		// Recover from panic if one occurred. Log the error and exit.
		if r := recover(); r != nil {
			log.Fatal().
				Str("error", fmt.Sprintf("%v", r)).
				Bytes("stack", debug.Stack()).
				Msg("Fatal error encountered")
		}
	}()

	err := godotenv.Load("agent.env")
	if err != nil {
		log.Error().Err(err).Msg("Error loading agent.env file")
	}

	cfg := config.NewConfigBuilder(&log).
		FromEnv().
		FromFlags().Build()

	log.Info().Interface("config", cfg).Msg("Configuration loaded")

	// WaitGroup to wait for all goroutines to complete
	var wg sync.WaitGroup

	ctx, _ := utils.WithSignalCancel(context.Background(), &log)

	log.Info().
		Str("server_address", cfg.Address).
		Msg("Endpoint address set")

	metricsStream := make(chan []db.Metric, domain.ChannelLength)

	var collector watcher.Collector = watcher.NewMetricCollector(metricsStream, cfg, &log)
	var reporter watcher.Reporter = watcher.NewMetricReporter(metricsStream, cfg, &log)

	log.Info().
		Int64("polling_time", cfg.PollInterval).
		Int64("reporting_time", cfg.ReportInterval).
		Msg("Polling and reporting times set")

	// Start watcher collection
	wg.Add(1)
	go collector.StartCollector(ctx, &wg)

	// Start watcher reporting
	wg.Add(1)
	go reporter.StartReporter(ctx, &wg)

	log.Info().Msg("Application started")
	utils.WaitForShutdown(metricsStream, &wg)
	log.Info().Msg("Application stopped gracefully")
}
