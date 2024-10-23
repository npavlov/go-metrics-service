package config_test

import (
	"flag"
	"os"
	"testing"

	testutils "github.com/npavlov/go-metrics-service/internal/test_utils"

	"github.com/caarlos0/env/v6"
	"github.com/npavlov/go-metrics-service/internal/agent/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewConfigBuilder checks if the default values are initialized properly.
func TestNewConfigBuilder(t *testing.T) {
	t.Parallel()

	l := testutils.GetTLogger()
	builder := config.NewConfigBuilder(l)
	assert.NotNil(t, builder.Build(), "Config should be initialized")
}

// TestFromEnv checks if environment variables are properly parsed into the config.
func TestFromEnv(t *testing.T) {
	// Set environment variables to test parsing
	t.Setenv("ADDRESS", "localhost:8080")
	t.Setenv("REPORT_INTERVAL", "10")
	t.Setenv("POLL_INTERVAL", "5")

	l := testutils.GetTLogger()
	cfg := config.NewConfigBuilder(l).FromEnv().Build()

	// Manually parse the environment variables to a temporary config for comparison
	tmpConfig := &config.Config{}
	err := env.Parse(tmpConfig)
	require.NoError(t, err, "Env parsing should not produce an error")

	assert.Equal(t, "http://"+tmpConfig.Address, cfg.Address, "Address should match the env value")
	assert.Equal(t, tmpConfig.ReportInterval, cfg.ReportInterval, "ReportInterval should match the env value")
	assert.Equal(t, tmpConfig.PollInterval, cfg.PollInterval, "PollInterval should match the env value")
}

// TestFromFlags checks if command line flags are properly parsed into the config.
func TestFromFlags(t *testing.T) {
	t.Parallel()
	// Reset command-line flags between tests
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	// Prepare the command-line arguments to test
	os.Args = []string{
		"cmd",
		"-a", "localhost:8081",
		"-r", "20",
		"-p", "10",
	}

	l := testutils.GetTLogger()
	cfg := config.NewConfigBuilder(l).FromFlags().Build()

	// Verify that flags were correctly parsed into the config
	assert.Equal(t, "http://localhost:8081", cfg.Address, "Address should be set by flag")
	assert.Equal(t, int64(20), cfg.ReportInterval, "ReportInterval should be set by flag")
	assert.Equal(t, int64(10), cfg.PollInterval, "PollInterval should be set by flag")
}
