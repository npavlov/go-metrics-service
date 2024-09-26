package main

import (
	"context"
	"fmt"
	metrics "github.com/npavlov/go-metrics-service/internal/agent/metrics-service"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	var service metrics.Service = metrics.NewMetricService()

	fmt.Println(service)

	pollInterval := 2 * time.Second
	reportInterval := 10 * time.Second

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
				time.Sleep(pollInterval)
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
				time.Sleep(reportInterval)
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
