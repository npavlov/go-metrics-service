package storage

import (
	"context"
	"hash/fnv"

	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

// AcquireBlockingLock attempts to acquire an advisory lock using two int32 keys.
func AcquireBlockingLock(ctx context.Context, query pgx.Tx, lockKey1, lockKey2 int32, log *zerolog.Logger) error {
	// Use Exec to execute the blocking lock function pg_advisory_xact_lock
	_, err := query.Exec(ctx, "SELECT pg_advisory_xact_lock($1, $2)", lockKey1, lockKey2)
	if err != nil {
		log.Error().Msg("failed to acquire blocking lock for metric")

		return errors.Wrapf(err, "failed to acquire blocking lock for metric with keys %d, %d", lockKey1, lockKey2)
	}

	log.Info().Msg("successfully acquired blocking lock for metric")

	return nil
}

// KeyNameAsHash64 converts a key name into two int32 values for advisory locking.
func KeyNameAsHash64(keyName string) (int32, int32) {
	hash := fnv.New64()
	_, err := hash.Write([]byte(keyName))
	if err != nil {
		panic(err)
	}
	hashValue := hash.Sum64()

	// Split the 64-bit hash into two 32-bit integers
	//nolint:gosec,mnd
	return int32(hashValue >> 32), int32(hashValue & 0xFFFFFFFF)
}
