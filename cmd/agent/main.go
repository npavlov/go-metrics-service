package main

import (
	"context"
	"fmt"
	"github.com/npavlov/go-metrics-service/internal/agent/metrics"
	"github.com/npavlov/go-metrics-service/internal/flags"
	"github.com/npavlov/go-metrics-service/internal/storage"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	parseFlags()
	flags.VerifyFlags()

	var memStorage storage.Repository = storage.NewMemStorage()

	fmt.Printf("Enpoint address set as %s\n", flagRunAddr)
	var service metrics.Service = metrics.NewMetricService(memStorage, "http://"+flagRunAddr)

	fmt.Printf("Polling time time %d, reporint time %d\n", pollInterval, reportInterval)

	// Create a context that will be canceled when a shutdown signal is received
	ctx, cancel := context.WithCancel(context.Background())

	// Channel to listen for OS signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start metrics collection
	go func() {
		for {
			select {
			case <-ctx.Done():
				fmt.Println("Stopping metrics collection")
				return
			default:
				service.UpdateMetrics()
				time.Sleep(pollInterval * time.Second)
			}
		}
	}()

	// Start metrics reporting
	go func() {
		for {
			select {
			case <-ctx.Done():
				fmt.Println("Stopping metrics reporting")
				return
			default:
				service.SendMetrics()
				time.Sleep(reportInterval * time.Second)
			}
		}
	}()

	// Wait for shutdown signal
	<-sigChan
	fmt.Println("Shutdown signal received")

	// Cancel the context to stop goroutines
	cancel()

	// Optionally, you can add cleanup logic here
	fmt.Println("Shutting down gracefully")
}
