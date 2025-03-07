package watcher

import "github.com/npavlov/go-metrics-service/internal/server/db"

type Result struct {
	Metric  *db.Metric  // Single metric (if applicable)
	Metrics []db.Metric // Array of stats (if applicable)
	Error   error       // Error (if any)
}
