package main

import (
	"context"
	"fmt"
	"github.com/npavlov/go-metrics-service/internal/agent/metrics"
	"github.com/npavlov/go-metrics-service/internal/flags"
	"github.com/npavlov/go-metrics-service/internal/shutdown"
	"github.com/npavlov/go-metrics-service/internal/storage"
)

func main() {
	cfg := parseFlags()
	flags.VerifyFlags()

	var memStorage storage.Repository = storage.NewMemStorage()

	fmt.Printf("Enpoint address set as %s\n", cfg.Address)
	var collector metrics.Collector = metrics.NewMetricCollector(memStorage)
	var reporter metrics.Reporter = metrics.NewMetricReporter(memStorage, "http://"+cfg.Address)

	fmt.Printf("Polling time time %d, reporint time %d\n", cfg.PollInterval, cfg.ReportInterval)

	ctx, cancel, sigChan := shutdown.WithSignalCancel(context.Background())

	// Start metrics collection
	go collector.StartCollector(ctx, cfg)

	// Start metrics reporting
	go reporter.StartReporter(ctx, cfg)

	shutdown.WaitForShutdown(sigChan, cancel)
	fmt.Println("Application stopped gracefully")
}
