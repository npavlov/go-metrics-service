package utils

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/rs/zerolog"

	"github.com/npavlov/go-metrics-service/internal/server/db"
)

func WithSignalCancel(ctx context.Context, log *zerolog.Logger) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(ctx)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Info().Msg("Shutdown signal received")
		cancel()
	}()

	return ctx, cancel
}

func WaitForShutdown(inputStream chan []db.Metric, wg *sync.WaitGroup) {
	sync.OnceFunc(func() {
		close(inputStream)
	})

	wg.Wait()
}
