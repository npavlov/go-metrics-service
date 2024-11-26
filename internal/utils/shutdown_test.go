package utils_test

import (
	"context"
	"os"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/npavlov/go-metrics-service/internal/server/db"

	testutils "github.com/npavlov/go-metrics-service/internal/test_utils"
	"github.com/npavlov/go-metrics-service/internal/utils"
)

func TestWithSignalCancel(t *testing.T) {
	t.Parallel()
	// Create a context and call WithSignalCancel
	l := testutils.GetTLogger()
	ctx := context.Background()
	ctxWithCancel, _ := utils.WithSignalCancel(ctx, l)

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
	require.NoError(t, err)

	// Send SIGINT signal to the current process
	err = process.Signal(syscall.SIGINT)
	require.NoError(t, err)

	// Wait for the goroutine to finish
	wg.Wait()

	// Check that the context is canceled
	assert.Equal(t, context.Canceled, ctxWithCancel.Err())
}

func TestWaitForShutdown(t *testing.T) {
	t.Parallel()

	var wg sync.WaitGroup

	// Simulate a task
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(100 * time.Millisecond) // Simulate some work
	}()

	// Call WaitForShutdown and check if it waits correctly
	start := time.Now()
	stream := make(chan []db.Metric, 1)
	utils.WaitForShutdown(stream, &wg)
	duration := time.Since(start)

	// Ensure that the WaitForShutdown finished after the simulated work
	require.GreaterOrEqual(t, duration, 100*time.Millisecond, "WaitForShutdown did not wait correctly")
}
