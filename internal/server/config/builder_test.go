package config_test

import (
	"flag"
	"os"
	"testing"

	"github.com/caarlos0/env/v6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/npavlov/go-metrics-service/internal/server/config"
	testutils "github.com/npavlov/go-metrics-service/internal/test_utils"
)

// TestNewConfigBuilder checks if the default values are initialized properly.
func TestNewConfigBuilder(t *testing.T) {
	t.Parallel()

	cfg := config.NewConfigBuilder(testutils.GetTLogger()).Build()
	assert.NotNil(t, cfg, "Config should be initialized")
}

// TestFromEnv checks if environment variables are properly parsed into the config.
func TestFromEnv(t *testing.T) {
	// Set environment variables to test parsing
	t.Setenv("ADDRESS", "localhost:8082")

	cfg := config.NewConfigBuilder(testutils.GetTLogger()).FromEnv().Build()

	// Manually parse the environment variables to a temporary config for comparison
	tmpConfig := &config.Config{}
	err := env.Parse(tmpConfig)
	require.NoError(t, err, "Env parsing should not produce an error")

	assert.Equal(t, tmpConfig.Address, cfg.Address, "Address should match the env value")
}

// TestFromFlags checks if command line flags are properly parsed into the config.
func TestFromFlags(t *testing.T) {
	t.Parallel()

	// Reset command-line flags between tests
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	// Prepare the command-line arguments to test
	os.Args = []string{
		"cmd",
		"-a", "localhost:8091",
	}

	cfg := config.NewConfigBuilder(testutils.GetTLogger()).FromFlags().Build()

	// Verify that flags were correctly parsed into the config
	assert.Equal(t, "localhost:8091", cfg.Address, "Address should be set by flag")
}

// TestFromFile checks if configuration is properly loaded from a file.
func TestFromFile(t *testing.T) {
	t.Parallel()

	l := testutils.GetTLogger()
	cfg := config.NewConfigBuilder(l).FromObj(&config.Config{
		Config: "testdata/config.json",
	}).FromFile().Build()

	// Verify that values were correctly loaded from the file
	assert.Equal(t, "localhost:9090", cfg.Address, "Address should be set by config file")
	assert.True(t, cfg.RestoreStorage, "RestoreStorage should be set by config file")
	assert.Equal(t, int64(1), cfg.StoreInterval, "PollInterval should be set by config file")
	assert.Equal(t, "/path/to/file.db", cfg.File, "File should be set by config file")
	assert.Equal(t, "postgres://", cfg.Database, "Database should be set by config file")
	assert.Equal(t, "./private.key", cfg.CryptoKey, "CryptoKey should be set by config file")
}
