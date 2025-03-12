package utils_test

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/npavlov/go-metrics-service/internal/utils"
)

func TestCalculateHash(t *testing.T) {
	t.Parallel()

	key := "secret"
	payload := []byte("data")
	expectedHMAC := hmac.New(sha256.New, []byte(key))
	expectedHMAC.Write(payload)
	expectedHash := hex.EncodeToString(expectedHMAC.Sum(nil))

	actualHash := utils.CalculateHash(key, payload)
	assert.Equal(t, expectedHash, actualHash)
}
