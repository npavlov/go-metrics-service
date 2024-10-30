package handlers_test

import (
	"context"
	"errors"
	"testing"

	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/npavlov/go-metrics-service/internal/server/repository"
)

func TestDBRepository_Ping(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		mockError   error
		expectError bool
	}{
		{
			name:        "Ping successful",
			mockError:   nil,
			expectError: false,
		},
		{
			name: "Ping failed",
			//nolint:err113
			mockError:   errors.New("failed to ping database"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Create a new pgxmock pool
			mockPool, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mockPool.Close()

			// Set up mock expectation for Ping
			if tt.mockError != nil {
				mockPool.ExpectPing().WillReturnError(tt.mockError)
			} else {
				mockPool.ExpectPing().WillReturnError(nil)
			}

			// Create a DBRepository with the mock pool
			repo := repository.NewDBRepository(mockPool)

			// Call the Ping method
			err = repo.Ping(context.Background())

			// Assert the expected error result based on test case
			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.mockError.Error())
			} else {
				require.NoError(t, err)
			}

			// Ensure all expectations were met
			require.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}
