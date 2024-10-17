package utils

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/npavlov/go-metrics-service/internal/logger"
)

func WithSignalCancel(ctx context.Context) context.Context {
	ctx, cancel := context.WithCancel(ctx)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	l := logger.NewLogger().Get()

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
