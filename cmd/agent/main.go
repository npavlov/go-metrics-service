package main

import (
	"context"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/npavlov/go-metrics-service/internal/agent/config"
	"github.com/npavlov/go-metrics-service/internal/agent/stats"
	"github.com/npavlov/go-metrics-service/internal/agent/watcher"
	"github.com/npavlov/go-metrics-service/internal/utils"
	"log"
	"sync"
)

func main() {
	err := godotenv.Load("agent.env")
	if err != nil {
		log.Fatal("Error loading agent.env file")
	}

	cfg := config.NewConfigBuilder().
		FromEnv().
		FromFlags().Build()

	metrics := stats.NewStats().StatsToMetrics()
	mux := sync.RWMutex{}

	fmt.Printf("Enpoint address set as %s\n", cfg.Address)
	var collector watcher.Collector = watcher.NewMetricCollector(&metrics, &mux)
	var reporter watcher.Reporter = watcher.NewMetricReporter(&metrics, &mux, "http://"+cfg.Address)

	fmt.Printf("Polling time time %d, reporint time %d\n", cfg.PollInterval, cfg.ReportInterval)

	ctx, cancel, sigChan := utils.WithSignalCancel(context.Background())

	// Start watcher collection
	go collector.StartCollector(ctx, cfg)

	// Start watcher reporting
	go reporter.StartReporter(ctx, cfg)

	utils.WaitForShutdown(sigChan, cancel)
	fmt.Println("Application stopped gracefully")
}
