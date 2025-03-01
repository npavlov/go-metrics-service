package grpc

import (
	"context"
	"net"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"

	"github.com/npavlov/go-metrics-service/internal/domain"
	"github.com/npavlov/go-metrics-service/internal/model"
	"github.com/npavlov/go-metrics-service/internal/server/config"
	"github.com/npavlov/go-metrics-service/internal/server/db"
	"github.com/npavlov/go-metrics-service/internal/utils"
	pb "github.com/npavlov/go-metrics-service/proto/v1"
)

type Server struct {
	pb.UnimplementedMetricServiceServer
	repo    model.Repository // Repository for accessing metric data.
	logger  *zerolog.Logger  // Logger for logging errors and info.
	cfg     *config.Config
	gServer *grpc.Server
}

func NewGRPCServer(repo model.Repository, cfg *config.Config, logger *zerolog.Logger) *Server {
	//nolint:exhaustruct
	return &Server{
		repo:   repo,
		logger: logger,
		cfg:    cfg,
		gServer: grpc.NewServer(grpc.ChainUnaryInterceptor(
			LoggingServerInterceptor(logger), // Logs all requests/responses
			SubnetInterceptor(cfg.TrustedSubnet, logger),
			SigInterceptor(cfg.Key, logger),
		)),
	}
}

func (gs *Server) Start(ctx context.Context) {
	go func() {
		gs.logger.Info().Str("address", gs.cfg.GRPCAddress).Msg("starting gRPC server")

		tcpListen, err := net.Listen("tcp", gs.cfg.GRPCAddress)
		if err != nil {
			gs.logger.Fatal().Err(err).Str("address", gs.cfg.GRPCAddress).Msg("failed to listen")
		}

		pb.RegisterMetricServiceServer(gs.gServer, gs)
		if err := gs.gServer.Serve(tcpListen); err != nil {
			gs.logger.Fatal().Err(err).Msg("failed to start gRPC server")
		}
	}()

	go func() {
		<-ctx.Done()
		gs.logger.Info().Msg("shutting down gRPC server")
		gs.gServer.GracefulStop()
	}()
}

func (gs *Server) SetMetrics(
	ctx context.Context,
	in *pb.MetricsRequest,
) (*pb.MetricsResponse, error) {
	newMetrics := make([]*db.Metric, 0, len(in.GetItems()))

	for _, metric := range in.GetItems() {
		newMetrics = append(newMetrics, utils.FromGModelToDBModel(metric))
	}

	// Collect metric IDs for database retrieval
	metricIDs := make([]domain.MetricName, len(in.GetItems()))
	for i, metric := range in.GetItems() {
		metricIDs[i] = domain.MetricName(metric.GetId())
	}

	// Fetch old metrics for updating existing ones
	oldMetrics, err := gs.repo.GetMany(ctx, metricIDs)
	if err != nil {
		gs.logger.Error().Err(err).Msg("error getting old metrics")

		return nil, errors.Wrap(err, "error getting old metrics")
	}

	// Prepare new metrics by updating existing ones or creating new entries
	for _, metric := range newMetrics {
		if oldMetric, found := oldMetrics[metric.ID]; found {
			oldMetric.SetValue(metric.Delta, metric.Value)
			oldMetrics[metric.ID] = oldMetric
		} else {
			oldMetrics[metric.ID] = *metric
		}
	}

	// Prepare the newMetrics slice with the updated oldMetrics
	newDBMetrics := make([]db.Metric, 0, len(oldMetrics))

	// Add all metrics from oldMetrics to newMetrics
	for _, oldMetric := range oldMetrics {
		newDBMetrics = append(newDBMetrics, oldMetric)
	}

	// Update all metrics in the repository
	if err = gs.repo.UpdateMany(ctx, &newDBMetrics); err != nil {
		return nil, errors.Wrap(err, "error updating old metrics")
	}

	newGRPcMetrics := make([]*pb.Metric, 0, len(newDBMetrics))
	for _, metric := range newDBMetrics {
		newGRPcMetrics = append(newGRPcMetrics, utils.FromDBModelToGModel(&metric))
	}

	return &pb.MetricsResponse{
		Status: true,
		Items:  newGRPcMetrics,
	}, nil
}

func (gs *Server) SetMetric(
	ctx context.Context,
	in *pb.MetricRequest,
) (*pb.MetricResponse, error) {
	newMetric := in.GetMetric()

	existingMetric, found := gs.repo.Get(ctx, domain.MetricName(newMetric.GetId()))

	if found {
		existingMetric.SetValue(&newMetric.Delta, &newMetric.Value)

		err := gs.repo.Update(ctx, existingMetric)
		if err != nil {
			return nil, errors.Wrap(err, "error updating existingMetric")
		}

		grpcModel := utils.FromDBModelToGModel(existingMetric)

		return &pb.MetricResponse{
			Status: true,
			Metric: grpcModel,
		}, nil
	}

	dbMetric := utils.FromGModelToDBModel(newMetric)

	err := gs.repo.Create(ctx, dbMetric)
	if err != nil {
		return nil, errors.Wrap(err, "error creating newMetric")
	}

	return &pb.MetricResponse{
		Status: true,
		Metric: newMetric,
	}, nil
}
