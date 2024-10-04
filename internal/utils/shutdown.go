package utils

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func WithSignalCancel(ctx context.Context) (context.Context, context.CancelFunc, chan os.Signal) {
	ctx, cancel := context.WithCancel(ctx)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		cancel()
	}()

	return ctx, cancel, sigChan
}

func WaitForShutdown(sigChan chan os.Signal, cancelFunc context.CancelFunc) {
	<-sigChan
	fmt.Println("Shutdown signal received")
	cancelFunc()
}
