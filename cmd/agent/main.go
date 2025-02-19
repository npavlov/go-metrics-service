package main

import (
	"context"
	"fmt"
	"runtime/debug"
	"sync"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog"

	"github.com/npavlov/go-metrics-service/internal/agent/buildinfo"
	"github.com/npavlov/go-metrics-service/internal/agent/config"
	"github.com/npavlov/go-metrics-service/internal/agent/watcher"
	"github.com/npavlov/go-metrics-service/internal/domain"
	"github.com/npavlov/go-metrics-service/internal/logger"
	"github.com/npavlov/go-metrics-service/internal/server/db"
	"github.com/npavlov/go-metrics-service/internal/utils"
)

func main() {
	log := logger.NewLogger(zerolog.DebugLevel).Get()

	log.Info().Str("buildVersion", buildinfo.Version).
		Str("buildCommit", buildinfo.Commit).
		Str("buildDate", buildinfo.Date).Msg("Starting agent")

	defer handlePanic(&log)

	cfg := loadConfig(&log)

	ctx, cancel := utils.WithSignalCancel(context.Background(), &log)
	defer cancel()

	runAgent(ctx, cfg, &log)
}

func handlePanic(log *zerolog.Logger) {
	if r := recover(); r != nil {
		log.Fatal().
			Str("error", fmt.Sprintf("%v", r)).
			Bytes("stack", debug.Stack()).
			Msg("Fatal error encountered")
	}
}

func loadConfig(log *zerolog.Logger) *config.Config {
	err := godotenv.Load("agent.env")
	if err != nil {
		log.Error().Err(err).Msg("Error loading agent.env file")
	}

	cfg := config.NewConfigBuilder(log).
		FromEnv().
		FromFlags().
		FromFile().Build()

	log.Info().Interface("config", cfg).Msg("Configuration loaded")

	return cfg
}

func runAgent(ctx context.Context, cfg *config.Config, log *zerolog.Logger) {
	log.Info().
		Str("server_address", cfg.Address).
		Msg("Endpoint address set")

	metricsStream := make(chan []db.Metric, domain.ChannelLength)
	var wg sync.WaitGroup

	collector := watcher.NewMetricCollector(metricsStream, cfg, log)
	reporter := watcher.NewMetricReporter(metricsStream, cfg, log)

	log.Info().
		Int64("polling_time", cfg.PollInterval).
		Int64("reporting_time", cfg.ReportInterval).
		Msg("Polling and reporting times set")

	wg.Add(1)
	go collector.StartCollector(ctx, &wg)

	wg.Add(1)
	go reporter.StartReporter(ctx, &wg)

	log.Info().Msg("Application started")
	utils.WaitForShutdown(metricsStream, &wg)
	log.Info().Msg("Application stopped gracefully")
}
