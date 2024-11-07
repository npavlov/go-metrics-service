package snapshot

import (
	"encoding/json"
	"io/fs"
	"os"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	"github.com/npavlov/go-metrics-service/internal/domain"
	"github.com/npavlov/go-metrics-service/internal/server/db"
)

type Snapshot interface {
	Save(state map[domain.MetricName]db.Metric) error
	Restore() (map[domain.MetricName]db.Metric, error)
}

// MemSnapshot encapsulates saving and restoring state of MemStorage.
type MemSnapshot struct {
	filePath string
	l        *zerolog.Logger
}

// NewMemSnapshot creates a new Memento with the given file path and logger.
func NewMemSnapshot(filePath string, l *zerolog.Logger) *MemSnapshot {
	return &MemSnapshot{
		filePath: filePath,
		l:        l,
	}
}

// Save stores the state of the metrics to the configured file.
func (m *MemSnapshot) Save(state map[domain.MetricName]db.Metric) error {
	file, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return errors.Wrap(err, "failed to marshal output")
	}

	err = os.WriteFile(m.filePath, file, fs.ModePerm)
	if err != nil {
		return errors.Wrap(err, "failed to save output")
	}

	return nil
}

// Restore loads the state from the file into the provided map.
func (m *MemSnapshot) Restore() (map[domain.MetricName]db.Metric, error) {
	file, err := os.ReadFile(m.filePath)
	if err != nil {
		m.l.Error().Err(err).Msg("failed to load output")

		return nil, errors.Wrap(err, "failed to load output")
	}

	newStorage := make(map[domain.MetricName]db.Metric)
	err = json.Unmarshal(file, &newStorage)
	if err != nil {
		m.l.Error().Err(err).Msg("failed to unmarshal output")

		return nil, errors.Wrap(err, "failed to unmarshal output")
	}

	return newStorage, nil
}
