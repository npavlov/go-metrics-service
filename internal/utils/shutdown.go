package utils

import (
	"context"
	"github.com/npavlov/go-metrics-service/internal/logger"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func WithSignalCancel(ctx context.Context) context.Context {
	ctx, cancel := context.WithCancel(ctx)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	l := logger.Get()

	go func() {
		<-sigChan
		l.Info().Msg("Shutdown signal received")
		cancel()
	}()

	return ctx
}

func WaitForShutdown(wg *sync.WaitGroup) {
	wg.Wait()
}
