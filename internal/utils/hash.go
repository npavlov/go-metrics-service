package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

func CalculateHash(key string, payload []byte) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write(payload)

	return hex.EncodeToString(h.Sum(nil))
}
