package config

import (
	"flag"
	"os"
	"testing"

	"github.com/caarlos0/env/v6"
	"github.com/stretchr/testify/assert"
)

// TestNewConfigBuilder checks if the default values are initialized properly
func TestNewConfigBuilder(t *testing.T) {
	builder := NewConfigBuilder()
	assert.NotNil(t, builder.cfg, "Config should be initialized")
}

// TestFromEnv checks if environment variables are properly parsed into the config
func TestFromEnv(t *testing.T) {
	// Set environment variables to test parsing
	_ = os.Setenv("ADDRESS", "localhost:8080")
	_ = os.Setenv("REPORT_INTERVAL", "10")
	_ = os.Setenv("POLL_INTERVAL", "5")

	defer func() {
		_ = os.Unsetenv("ADDRESS")
	}()
	defer func() {
		_ = os.Unsetenv("REPORT_INTERVAL")
	}()
	defer func() {
		_ = os.Unsetenv("POLL_INTERVAL")
	}()

	builder := NewConfigBuilder().FromEnv()

	// Manually parse the environment variables to a temporary config for comparison
	tmpConfig := &Config{}
	err := env.Parse(tmpConfig)
	assert.NoError(t, err, "Env parsing should not produce an error")

	assert.Equal(t, tmpConfig.Address, builder.cfg.Address, "Address should match the env value")
}

// TestFromFlags checks if command line flags are properly parsed into the config
func TestFromFlags(t *testing.T) {
	// Reset command-line flags between tests
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	// Prepare the command-line arguments to test
	os.Args = []string{
		"cmd",
		"-a", "localhost:8081",
	}

	builder := NewConfigBuilder().FromFlags()

	// Verify that flags were correctly parsed into the config
	assert.Equal(t, "localhost:8081", builder.cfg.Address, "Address should be set by flag")
}

// TestBuild checks if the Build function returns the final config
func TestBuild(t *testing.T) {
	builder := NewConfigBuilder()
	finalConfig := builder.Build()

	assert.NotNil(t, finalConfig, "Final config should not be nil")
	assert.Equal(t, builder.cfg, finalConfig, "The returned config should match the builder's config")
}
