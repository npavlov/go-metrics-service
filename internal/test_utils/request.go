package testutils

import (
	"net/http"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func SendServerRequest(t *testing.T, url string, route string, statusCode int) {
	t.Helper()

	req := resty.New().R()
	req.Method = http.MethodPost
	req.URL = url + route

	res, err := req.Send()

	require.NoError(t, err, "error making HTTP request")
	assert.Equal(t, statusCode, res.StatusCode())
}
