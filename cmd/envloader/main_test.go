package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestSplitEnv tests the splitEnv function.
func TestSplitEnv(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input    string
		expected []string
	}{
		{"KEY=VALUE", []string{"KEY", "VALUE"}},
		{"FOO=BAR=BAZ", []string{"FOO", "BAR=BAZ"}}, // First '=' should split
		{"EMPTY=", []string{"EMPTY", ""}},
		{"NOEQUALS", []string{"NOEQUALS", ""}}, // No '=' in input
	}

	for _, test := range tests {
		result := splitEnv(test.input)
		assert.Equal(t, test.expected, result, "Failed for input: "+test.input)
	}
}

// TestGetEnvAsMap tests getEnvAsMap function with a controlled environment.
func TestGetEnvAsMap(t *testing.T) {
	// Backup current environment
	oldEnv := os.Environ()

	// Set up test environment variables
	os.Clearenv()
	t.Setenv("TEST_KEY", "TEST_VALUE")
	t.Setenv("FOO", "BAR")

	// Call function
	envMap := getEnvAsMap()

	// Restore old environment
	for _, e := range oldEnv {
		pair := splitEnv(e)
		_ = os.Setenv(pair[0], pair[1])
	}

	// Assertions
	assert.Equal(t, "TEST_VALUE", envMap["TEST_KEY"])
	assert.Equal(t, "BAR", envMap["FOO"])
	assert.NotEmpty(t, envMap)
}
