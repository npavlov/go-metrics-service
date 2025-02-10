package storage_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/npavlov/go-metrics-service/internal/server/storage"
)

// Test case 1: Success case where transaction commits successfully.
func TestWithTx_Success(t *testing.T) {
	t.Parallel()

	_, mockDB := setupDBStorage(t)

	mockDB.ExpectBegin()

	mockDB.ExpectCommit()

	// Define the function to be passed to WithTx
	fn := func(_ context.Context, _ pgx.Tx) error {
		// Your logic for the transaction goes here, in this case, no-op
		return nil
	}

	// Call WithTx and assert no errors occurred
	err := storage.WithTx(context.Background(), mockDB, fn)
	require.NoError(t, err)
}

// Test case 2: Case where an error occurs inside the transaction function.
func TestWithTx_ErrorInTx(t *testing.T) {
	t.Parallel()

	_, mockDB := setupDBStorage(t)

	mockDB.ExpectBegin()

	mockDB.ExpectRollback()

	// Define the function to be passed to WithTx which will return an error
	fn := func(_ context.Context, _ pgx.Tx) error {
		//nolint:err113
		return errors.New("some error")
	}

	// Call WithTx and assert that an error occurred
	err := storage.WithTx(context.Background(), mockDB, fn)
	require.Error(t, err)
	assert.Equal(t, "Tx.WithTxFunc: some error", err.Error())
}

// Test case 3: Case where Commit fails.
func TestWithTx_CommitError(t *testing.T) {
	t.Parallel()

	_, mockDB := setupDBStorage(t)

	mockDB.ExpectBegin()
	mockDB.ExpectCommit()
	mockDB.ExpectRollback()

	// Define the function to be passed to WithTx
	fn := func(ctx context.Context, tx pgx.Tx) error {
		_ = tx.Commit(ctx)
		//nolint:err113
		return errors.New("commit error")
	}

	// Call WithTx and assert that an error occurred
	err := storage.WithTx(context.Background(), mockDB, fn)
	require.Error(t, err)
	assert.Equal(t, "Tx.WithTxFunc: commit error", err.Error())
}
