package middlewares_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/npavlov/go-metrics-service/internal/server/middlewares"
)

// TestTimeoutMiddleware_Success ensures that the middleware works correctly within the allowed time limit.
func TestTimeoutMiddleware_Success(t *testing.T) {
	t.Parallel()

	timeout := 2 * time.Second
	handler := middlewares.TimeoutMiddleware(timeout)(http.HandlerFunc(func(writer http.ResponseWriter, r *http.Request) {
		// Check if context deadline is set
		deadline, ok := r.Context().Deadline()
		assert.True(t, ok, "Expected context to have a deadline")
		assert.WithinDuration(t, time.Now().Add(timeout), deadline, time.Second, "Expected context deadline to be close to the timeout")

		writer.WriteHeader(http.StatusOK)
	}))

	// Create an HTTP request to pass to the handler
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// Assert status code is 200 OK
	assert.Equal(t, http.StatusOK, rec.Code, "Expected status code 200 OK")
}

// TestTimeoutMiddleware_TimeoutExceeded ensures that the middleware cancels the context if the timeout is exceeded.
func TestTimeoutMiddleware_TimeoutExceeded(t *testing.T) {
	t.Parallel()

	timeout := 100 * time.Millisecond
	handler := middlewares.TimeoutMiddleware(timeout)(http.HandlerFunc(func(writer http.ResponseWriter, r *http.Request) {
		// Simulate a long-running process that exceeds the timeout
		select {
		case <-r.Context().Done():
			// If the context is canceled, we should see an error in the context
			assert.Equal(t, context.DeadlineExceeded, r.Context().Err(), "Expected context error to be DeadlineExceeded")
			http.Error(writer, "Request timed out", http.StatusGatewayTimeout)
		case <-time.After(500 * time.Millisecond):
			// This should not happen as the timeout will cancel the context first
			writer.WriteHeader(http.StatusOK)
		}
	}))

	// Create an HTTP request to pass to the handler
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// Assert status code is 504 Gateway Timeout
	assert.Equal(t, http.StatusGatewayTimeout, rec.Code, "Expected status code 504 Gateway Timeout")
	assert.Contains(t, rec.Body.String(), "Request timed out", "Expected timeout message in response")
}
