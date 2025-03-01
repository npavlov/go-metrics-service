package watcher

import (
	"context"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	"github.com/npavlov/go-metrics-service/internal/agent/config"
	au "github.com/npavlov/go-metrics-service/internal/agent/utils"
	"github.com/npavlov/go-metrics-service/internal/server/db"
	"github.com/npavlov/go-metrics-service/internal/utils"
	pb "github.com/npavlov/go-metrics-service/proto/v1"
)

type GRPCSender struct {
	conn   *grpc.ClientConn
	client pb.MetricServiceClient
	logger *zerolog.Logger
}

func NewGRPCSender(cfg *config.Config, logger *zerolog.Logger) *GRPCSender {
	ip := au.GetLocalIP(logger)

	interceptor := grpc.WithUnaryInterceptor(makeInterceptor(cfg, ip, logger))

	conn, err := grpc.NewClient(cfg.GRPCAddress, grpc.WithTransportCredentials(insecure.NewCredentials()), interceptor)
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

// makeInterceptor adds X-Real-IP and HashSHA256 metadata.
func makeInterceptor(cfg *config.Config, ip string, logger *zerolog.Logger) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string, req,
		reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		md := metadata.New(map[string]string{
			"X-Real-IP": ip,
		})

		// Serialize request and calculate hash if key is set
		if cfg.Key != "" {
			payload, err := utils.MarshalProtoMessage(req)
			if err != nil {
				logger.Error().Err(err).Msg("failed to marshal request for hashing")
			} else {
				hash := utils.CalculateHash(cfg.Key, payload)
				md.Append("HashSHA256", hash)
			}
		}

		ctx = metadata.NewOutgoingContext(ctx, md)

		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func (gc *GRPCSender) Close() {
	if err := gc.conn.Close(); err != nil {
		gc.logger.Error().Err(err).Msg("failed to close grpc connection")
	}
}

func (gc *GRPCSender) SendMetricsBatch(ctx context.Context, metrics []db.Metric) ([]db.Metric, error) {
	request := pb.MetricsRequest{
		Items: make([]*pb.Metric, len(metrics)),
	}
	for _, m := range metrics {
		request.Items = append(request.Items, utils.FromDBModelToGModel(&m))
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
	request := pb.MetricRequest{
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
