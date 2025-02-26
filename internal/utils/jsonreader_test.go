package utils_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	testutils "github.com/npavlov/go-metrics-service/internal/test_utils"
	"github.com/npavlov/go-metrics-service/internal/utils"
)

// TestReadFromFile verifies file loading.
func TestReadFromFile(t *testing.T) {
	t.Parallel()

	logger := testutils.GetTLogger()

	type Config struct {
		Address string `json:"address"`
		Restore bool   `json:"restore"`
	}

	cfg := &Config{}

	err := utils.ReadFromFile("testdata/test.json", cfg, logger)

	require.NoError(t, err)

	assert.Equal(t, "localhost:8080", cfg.Address)
	assert.True(t, cfg.Restore)
}

// TestReplaceValues verifies value copying.
func TestReplaceValues(t *testing.T) {
	t.Parallel()

	type Config struct {
		Address string
		Restore bool
	}

	src := &Config{Address: "localhost:8080", Restore: true}
	tgt := &Config{Address: "", Restore: false}

	utils.ReplaceValues(src, tgt)

	assert.Equal(t, "localhost:8080", tgt.Address)
	assert.True(t, tgt.Restore)
}
