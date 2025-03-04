package grpcsender

import (
	"context"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/npavlov/go-metrics-service/gen/go/proto/metrics/v1"
	"github.com/npavlov/go-metrics-service/internal/agent/config"
	au "github.com/npavlov/go-metrics-service/internal/agent/utils"
	"github.com/npavlov/go-metrics-service/internal/server/db"
	"github.com/npavlov/go-metrics-service/internal/utils"
	"github.com/npavlov/go-metrics-service/pkg/crypto"
)

type GRPCSender struct {
	conn   *grpc.ClientConn
	client pb.MetricServiceClient
	logger *zerolog.Logger
}

func NewGRPCSender(cfg *config.Config, logger *zerolog.Logger) *GRPCSender {
	ip := au.GetLocalIP(logger)

	// interceptor := grpc.WithUnaryInterceptor(HeadersInterceptor(cfg, ip, logger))

	var encryption *crypto.Encryption

	if key := cfg.CryptoKey; key != "" {
		var err error
		if encryption, err = crypto.NewEncryption(key); err != nil {
			logger.Fatal().Err(err).Msg("failed to create encryption")
		}
	}

	interceptors := grpc.WithChainUnaryInterceptor(
		HeadersInterceptor(cfg, ip, logger),
		EncodingInterceptor(encryption, logger),
	)

	conn, err := grpc.NewClient(cfg.GRPCAddress, grpc.WithTransportCredentials(insecure.NewCredentials()), interceptors)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to create grpc connection")
	}

	client := pb.NewMetricServiceClient(conn)

	return &GRPCSender{
		conn:   conn,
		client: client,
		logger: logger,
	}
}

func (gc *GRPCSender) Close() {
	if err := gc.conn.Close(); err != nil {
		gc.logger.Error().Err(err).Msg("failed to close grpc connection")
	}
}

func (gc *GRPCSender) SendMetricsBatch(ctx context.Context, metrics []db.Metric) ([]db.Metric, error) {
	//nolint:exhaustruct
	request := pb.SetMetricsRequest{
		Items: make([]*pb.Metric, len(metrics)),
	}
	for idx, m := range metrics {
		request.Items[idx] = utils.FromDBModelToGModel(&m)
	}

	resp, err := gc.client.SetMetrics(ctx, &request)
	if err != nil {
		gc.logger.Error().Err(err).Msg("failed to send metrics")

		return nil, errors.Wrap(err, "failed to send metrics using GRPC")
	}

	newMetrics := make([]db.Metric, len(resp.GetItems()))

	for i, item := range resp.GetItems() {
		newMetrics[i] = *utils.FromGModelToDBModel(item)
	}

	return newMetrics, nil
}

func (gc *GRPCSender) SendMetric(ctx context.Context, metric db.Metric) (*db.Metric, error) {
	//nolint:exhaustruct
	request := pb.SetMetricRequest{
		Metric: utils.FromDBModelToGModel(&metric),
	}

	resp, err := gc.client.SetMetric(ctx, &request)
	if err != nil {
		gc.logger.Error().Err(err).Msg("failed to send metrics")

		return nil, errors.Wrap(err, "failed to send metrics using GRPC")
	}

	newMetric := utils.FromGModelToDBModel(resp.GetMetric())

	return newMetric, nil
}
