package watcher

import (
	"context"
	"sync"
	"time"

	"github.com/rs/zerolog"

	"github.com/npavlov/go-metrics-service/internal/agent/config"
	"github.com/npavlov/go-metrics-service/internal/agent/model"
	"github.com/npavlov/go-metrics-service/internal/agent/watcher/grpcsender"
	"github.com/npavlov/go-metrics-service/internal/agent/watcher/jsonsender"
	"github.com/npavlov/go-metrics-service/internal/domain"
	"github.com/npavlov/go-metrics-service/internal/server/db"
)

// Reporter interface defines the contract for sending watcher.
type Reporter interface {
	StartReporter(ctx context.Context, wg *sync.WaitGroup)
}

// MetricReporter implements the Reporter interface.
type MetricReporter struct {
	cfg           *config.Config
	l             *zerolog.Logger
	workerCount   int
	inputStream   chan []db.Metric
	metricsStream chan db.Metric
	batchStream   chan []db.Metric
	resultStream  chan Result
	sender        model.Sender
}

func NewMetricReporter(inputStream chan []db.Metric, cfg *config.Config, logger *zerolog.Logger) *MetricReporter {
	reporter := &MetricReporter{
		cfg:           cfg,
		l:             logger,
		workerCount:   cfg.RateLimit,
		metricsStream: make(chan db.Metric, domain.ChannelLength),
		batchStream:   make(chan []db.Metric, domain.ChannelLength),
		resultStream:  make(chan Result),
		inputStream:   inputStream,
		sender:        nil,
	}

	// choose type of communication
	if cfg.UseGRPC {
		conn := grpcsender.MakeConnection(cfg, logger)
		reporter.sender = grpcsender.NewGRPCSender(conn, logger)
	} else {
		reporter.sender = jsonsender.NewSender(cfg, logger)
	}

	return reporter
}

func (mr *MetricReporter) StartReporter(ctx context.Context, wg *sync.WaitGroup) {
	// Start generator and worker pool
	go mr.metricGenerator(ctx, wg)

	workerWg := &sync.WaitGroup{}

	for i := range mr.workerCount {
		workerWg.Add(1)
		go mr.worker(ctx, workerWg, i)
	}

	go mr.ProcessResults(ctx)

	mr.l.Info().Msg("All workers and processors started")

	workerWg.Wait()
}

func (mr *MetricReporter) metricGenerator(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	// Buffer to hold incoming metrics
	var metricBuffer []db.Metric

	for {
		select {
		case <-ctx.Done():
			mr.l.Info().Msg("Stopping metric generator")
			mr.sender.Close()
			close(mr.metricsStream)
			close(mr.batchStream)

			return
		case inputData, ok := <-mr.inputStream: // Get metrics from the input channel
			if !ok {
				mr.l.Warn().Msg("Metrics channel closed, no more metrics to generate")

				return
			}
			metricBuffer = inputData
		default:
			mr.FillStream(metricBuffer)
			time.Sleep(mr.cfg.ReportIntervalDur)
		}
	}
}

func (mr *MetricReporter) FillStream(metrics []db.Metric) {
	if metrics == nil {
		return
	}

	if mr.cfg.UseBatch {
		mr.batchStream <- metrics
	} else {
		for _, metric := range metrics {
			mr.metricsStream <- metric
		}
	}
}

func (mr *MetricReporter) worker(ctx context.Context, wg *sync.WaitGroup, workerID int) {
	mr.l.Info().Int("worker Id", workerID).Msg("Worker has started")

	for {
		select {
		case <-ctx.Done():
			mr.l.Info().Int("WorkerId", workerID).Msg("Worker stopping due to context cancellation")
			wg.Done()

			return
		case metric, ok := <-mr.metricsStream:
			if !ok {
				mr.l.Info().Msg("Metrics channel closed, stopping worker")

				return
			}
			mr.l.Info().Int("worker Id", workerID).Msg("Worker is sending metric")
			result := mr.handleMetric(ctx, metric)
			mr.resultStream <- result
		case metrics, ok := <-mr.batchStream:
			if !ok {
				mr.l.Info().Msg("Batch channel closed, stopping worker")

				return
			}
			mr.l.Info().Int("worker Id", workerID).Msg("Worker is sending batch stats")
			result := mr.handleBatch(ctx, metrics)
			mr.resultStream <- result
		}
	}
}

func (mr *MetricReporter) handleMetric(ctx context.Context, metric db.Metric) Result {
	data, err := mr.sender.SendMetric(ctx, metric)

	return Result{
		Metric:  data,
		Error:   err,
		Metrics: nil,
	}
}

func (mr *MetricReporter) handleBatch(ctx context.Context, metrics []db.Metric) Result {
	data, err := mr.sender.SendMetricsBatch(ctx, metrics)

	return Result{
		Metrics: data,
		Error:   err,
		Metric:  nil,
	}
}

func (mr *MetricReporter) ProcessResults(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			mr.l.Info().Msg("Stopping result processor")

			return
		case result, ok := <-mr.resultStream:
			if !ok {
				mr.l.Info().Msg("Result channel closed, stopping processor")

				return
			}

			// Log or handle the result
			if result.Error != nil {
				mr.l.Error().Err(result.Error).Msg("Error sending metric")
			}
			if result.Metric != nil {
				mr.l.Info().Interface("metric", result.Metric).Msg("Processed single metric successfully")
			}
			if result.Metrics != nil {
				mr.l.Info().Interface("stats", result.Metrics).Msg("Processed batch stats successfully")
			}
		}
	}
}
