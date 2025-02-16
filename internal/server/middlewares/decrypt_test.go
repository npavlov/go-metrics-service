package middlewares_test

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/npavlov/go-metrics-service/pkg/crypto"

	"github.com/npavlov/go-metrics-service/internal/server/middlewares"
)

func TestDecryptMiddleware(t *testing.T) {
	t.Parallel()

	logger := zerolog.New(io.Discard) // Discard logs during tests

	decryption, err := crypto.NewDecryption("./testdata/test_private.key")
	require.NoError(t, err)
	encryption, err := crypto.NewEncryption("./testdata/test_public.key")
	require.NoError(t, err)
	middleware := middlewares.DecryptMiddleware(decryption, &logger)

	t.Run("should pass through if x-encrypted is not true", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodPost, "http://example.com", bytes.NewBufferString("test payload"))
		rec := httptest.NewRecorder()
		req.Header.Set("Content-Type", "application/json")

		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "test payload", readBody(r))
			w.WriteHeader(http.StatusOK)
		}))

		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("should decrypt body if x-encrypted is true", func(t *testing.T) {
		t.Parallel()

		payload := "encrypted data"
		encryptedPayload, err := encryption.Encrypt([]byte(payload))
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "http://example.com", bytes.NewBuffer(encryptedPayload))
		rec := httptest.NewRecorder()
		req.Header.Set("X-Encrypted", "true")

		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, payload, readBody(r))
			w.WriteHeader(http.StatusOK)
		}))

		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
	})
}

func readBody(r *http.Request) string {
	body, _ := io.ReadAll(r.Body)
	r.Body = io.NopCloser(bytes.NewBuffer(body)) // Restore the body for further reads if needed

	return string(body)
}
