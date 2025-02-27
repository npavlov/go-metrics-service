package utils_test

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	testutils "github.com/npavlov/go-metrics-service/internal/test_utils"

	"github.com/npavlov/go-metrics-service/internal/agent/utils"
)

func TestGetLocalIP(t *testing.T) {
	t.Parallel()

	logger := testutils.GetTLogger()

	ip := utils.GetLocalIP(logger)

	if ip == "" {
		t.Skip("Skipping test as no valid IP was found")
	}

	parsedIP := net.ParseIP(ip)
	require.NotNil(t, parsedIP, "Expected a valid IP address, but got nil")
	assert.False(t, parsedIP.IsLoopback(), "IP should not be a loopback address")
	assert.NotNil(t, parsedIP.To4(), "IP should be an IPv4 address")
}
