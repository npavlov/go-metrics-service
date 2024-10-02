package main

import (
	"context"
	"fmt"
	"github.com/npavlov/go-metrics-service/internal/agent/stats"
	"github.com/npavlov/go-metrics-service/internal/agent/watcher"
	"github.com/npavlov/go-metrics-service/internal/flags"
	"github.com/npavlov/go-metrics-service/internal/shutdown"
	"sync"
)

func main() {
	cfg := parseFlags()
	flags.VerifyFlags()

	st := stats.Stats{}
	metrics := st.StatsToMetrics()
	mux := sync.RWMutex{}

	fmt.Printf("Enpoint address set as %s\n", cfg.Address)
	var collector watcher.Collector = watcher.NewMetricCollector(&metrics, &mux)
	var reporter watcher.Reporter = watcher.NewMetricReporter(&metrics, &mux, "http://"+cfg.Address)

	fmt.Printf("Polling time time %d, reporint time %d\n", cfg.PollInterval, cfg.ReportInterval)

	ctx, cancel, sigChan := shutdown.WithSignalCancel(context.Background())

	// Start watcher collection
	go collector.StartCollector(ctx, cfg)

	// Start watcher reporting
	go reporter.StartReporter(ctx, cfg)

	shutdown.WaitForShutdown(sigChan, cancel)
	fmt.Println("Application stopped gracefully")
}
