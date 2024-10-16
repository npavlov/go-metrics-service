package utils

import (
	"context"
	"os"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithSignalCancel(t *testing.T) {
	// Create a context and call WithSignalCancel
	ctx := context.Background()
	ctxWithCancel := WithSignalCancel(ctx)

	// Create a wait group to wait for the cancellation
	var wg sync.WaitGroup
	wg.Add(1)

	// Launch a goroutine to wait for cancellation
	go func() {
		defer wg.Done()
		<-ctxWithCancel.Done()
	}()

	// Simulate sending SIGINT to the process
	process, err := os.FindProcess(os.Getpid()) // Get the current process
	assert.NoError(t, err)

	// Send SIGINT signal to the current process
	err = process.Signal(syscall.SIGINT)
	assert.NoError(t, err)

	// Wait for the goroutine to finish
	wg.Wait()

	// Check that the context is canceled
	assert.Equal(t, context.Canceled, ctxWithCancel.Err())
}

func TestWaitForShutdown(t *testing.T) {
	var wg sync.WaitGroup

	// Simulate a task
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(100 * time.Millisecond) // Simulate some work
	}()

	// Call WaitForShutdown and check if it waits correctly
	start := time.Now()
	WaitForShutdown(&wg)
	duration := time.Since(start)

	// Ensure that the WaitForShutdown finished after the simulated work
	require.True(t, duration >= 100*time.Millisecond, "WaitForShutdown did not wait correctly")
}
