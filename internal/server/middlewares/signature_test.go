package middlewares_test

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"

	"github.com/npavlov/go-metrics-service/internal/server/middlewares"
)

func TestSignatureMiddleware(t *testing.T) {
	t.Parallel()

	logger := zerolog.New(io.Discard) // Discard logger output in tests
	signKey := "test_secret_key"

	middleware := middlewares.SignatureMiddleware(signKey, &logger)

	// Helper to calculate valid HMAC signature for a payload
	calculateSignature := func(payload []byte, key string) string {
		h := hmac.New(sha256.New, []byte(key))
		h.Write(payload)

		return hex.EncodeToString(h.Sum(nil))
	}

	t.Run("Valid POST request with correct signature", func(t *testing.T) {
		t.Parallel()

		payload := []byte(`{"data":"test"}`)
		signature := calculateSignature(payload, signKey)

		// Create a request with a valid signature
		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(payload))
		req.Header.Set("HashSHA256", signature)

		// Recorder to capture the response
		rr := httptest.NewRecorder()

		// Mock next handler
		nextHandler := http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
			writer.WriteHeader(http.StatusOK)
		})

		// Call middleware
		middleware(nextHandler).ServeHTTP(rr, req)

		// Assert the response status
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("Invalid POST request with incorrect signature", func(t *testing.T) {
		t.Parallel()

		payload := []byte(`{"data":"test"}`)
		invalidSignature := "invalid_signature"

		// Create a request with an invalid signature
		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(payload))
		req.Header.Set("HashSHA256", invalidSignature)

		// Recorder to capture the response
		rr := httptest.NewRecorder()

		// Mock next handler
		nextHandler := http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
			writer.WriteHeader(http.StatusOK)
		})

		// Call middleware
		middleware(nextHandler).ServeHTTP(rr, req)

		// Assert the response status
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("POST request without signature", func(t *testing.T) {
		t.Parallel()

		payload := []byte(`{"data":"test"}`)

		// Create a request without the `HashSHA256` header
		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(payload))

		// Recorder to capture the response
		rr := httptest.NewRecorder()

		// Mock next handler
		nextHandler := http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
			writer.WriteHeader(http.StatusOK)
		})

		// Call middleware
		middleware(nextHandler).ServeHTTP(rr, req)

		// Assert the response status
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("Non-POST request", func(t *testing.T) {
		t.Parallel()

		// Create a GET request (not POST)
		req := httptest.NewRequest(http.MethodGet, "/", nil)

		// Recorder to capture the response
		rr := httptest.NewRecorder()

		// Mock next handler
		nextHandler := http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
			writer.WriteHeader(http.StatusOK)
		})

		// Call middleware
		middleware(nextHandler).ServeHTTP(rr, req)

		// Assert the response status
		assert.Equal(t, http.StatusOK, rr.Code)
	})
}