package grpc

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/bufbuild/protovalidate-go"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	pb "github.com/npavlov/go-metrics-service/gen/go/proto/metrics/v1"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/npavlov/go-metrics-service/internal/domain"
	"github.com/npavlov/go-metrics-service/internal/model"
	"github.com/npavlov/go-metrics-service/internal/server/config"
	"github.com/npavlov/go-metrics-service/internal/server/db"
	"github.com/npavlov/go-metrics-service/internal/utils"
	"github.com/npavlov/go-metrics-service/pkg/crypto"
)

type Server struct {
	pb.UnimplementedMetricServiceServer
	repo    model.Repository // Repository for accessing metric data.
	logger  *zerolog.Logger  // Logger for logging errors and info.
	cfg     *config.Config
	gServer *grpc.Server
}

func NewGRPCServer(repo model.Repository, cfg *config.Config, logger *zerolog.Logger) *Server {
	var decryption *crypto.Decryption

	if key := cfg.CryptoKey; key != "" {
		var err error
		if decryption, err = crypto.NewDecryption(key); err != nil {
			logger.Fatal().Err(err).Msg("failed to create decryption")
		}
	}

	//nolint:exhaustruct
	return &Server{
		repo:   repo,
		logger: logger,
		cfg:    cfg,
		gServer: grpc.NewServer(grpc.ChainUnaryInterceptor(
			LoggingServerInterceptor(logger), // Logs all requests/responses
			SubnetInterceptor(cfg.TrustedSubnet, logger),
			DecryptInterceptor(decryption, logger),
			SigInterceptor(cfg.Key, logger),
		)),
	}
}

func (gs *Server) Start(ctx context.Context) {
	// Start gRPC-server in goroutine
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

	// Start gRPC-Gateway in another goroutine
	go func() {
		mux := runtime.NewServeMux()
		opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

		// Register gRPC-Gateway handlers
		err := pb.RegisterMetricServiceHandlerFromEndpoint(ctx, mux, gs.cfg.GRPCAddress, opts)
		if err != nil {
			gs.logger.Fatal().Err(err).Msg("failed to register gRPC-Gateway")
		}

		//nolint:exhaustruct
		server := &http.Server{
			Addr:         gs.cfg.GRPCGateway,
			ReadTimeout:  1 * time.Second,
			WriteTimeout: 1 * time.Second,
			Handler:      mux,
		}

		gs.logger.Info().Str("address", gs.cfg.GRPCGateway).Msg("starting gRPC-Gateway")
		if err := server.ListenAndServe(); err != nil {
			gs.logger.Fatal().Err(err).Msg("failed to start gRPC-Gateway")
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
	in *pb.SetMetricsRequest,
) (*pb.SetMetricsResponse, error) {
	validator, err := protovalidate.New()
	if err != nil {
		return nil, errors.Wrap(err, "error initializing validator")
	}

	if err := validator.Validate(in); err != nil {
		return nil, errors.Wrap(err, "error validating input")
	}

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

	return &pb.SetMetricsResponse{
		Status: true,
		Items:  newGRPcMetrics,
	}, nil
}

func (gs *Server) SetMetric(
	ctx context.Context,
	in *pb.SetMetricRequest,
) (*pb.SetMetricResponse, error) {
	newMetric := in.GetMetric()

	validator, err := protovalidate.New()
	if err != nil {
		return nil, errors.Wrap(err, "error initializing validator")
	}

	if err := validator.Validate(in); err != nil {
		return nil, errors.Wrap(err, "error validating input")
	}

	existingMetric, found := gs.repo.Get(ctx, domain.MetricName(newMetric.GetId()))

	if found {
		existingMetric.SetValue(newMetric.Delta, newMetric.Value)

		err := gs.repo.Update(ctx, existingMetric)
		if err != nil {
			return nil, errors.Wrap(err, "error updating existingMetric")
		}

		grpcModel := utils.FromDBModelToGModel(existingMetric)

		return &pb.SetMetricResponse{
			Status: true,
			Metric: grpcModel,
		}, nil
	}

	dbMetric := utils.FromGModelToDBModel(newMetric)

	err = gs.repo.Create(ctx, dbMetric)
	if err != nil {
		return nil, errors.Wrap(err, "error creating newMetric")
	}

	return &pb.SetMetricResponse{
		Status: true,
		Metric: newMetric,
	}, nil
}
