package main

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5/pgxpool"
	myLogger "github.com/kdaadk/route256/pkg/logger"
	"github.com/kdaadk/route256/pkg/metrics"
	"github.com/kdaadk/route256/pkg/tracing"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"route256/loms/internal/handler"
	"route256/loms/internal/kafka"
	"route256/loms/internal/model"
	"route256/loms/internal/mw"
	"route256/loms/internal/repository"
	"route256/loms/internal/service"
	"route256/loms/proto"
	"syscall"
)

var serviceName = "loms-service"

func main() {
	// logging
	config := zap.NewDevelopmentConfig()
	config.ErrorOutputPaths = []string{"output"}
	config.Level.SetLevel(zapcore.InfoLevel)
	logger := myLogger.NewMyLogger(config)
	defer logger.Sync()

	//tracing
	tp, err := tracing.InitTracer("jaeger:4317", serviceName, tracing.GRPCTransport)
	if err != nil {
		myLogger.ErrorContext(context.Background(), "Failed to initialize tracer", zap.Error(err))
		return
	}
	defer tracing.ShutdownTracer(tp)

	tracer := tp.Tracer(serviceName)

	// postgres
	dbPool, err := initDBPools()
	if err != nil {
		myLogger.ErrorContext(context.Background(), "Error connecting to database", zap.Error(err))
		return
	}

	// server
	txManager := repository.NewTxManager(dbPool)
	ordersRepo := repository.NewOrdersRepository(dbPool)
	stocksRepo := repository.NewStocksRepository(dbPool)
	kafkaProducer, err := kafka.NewKafkaProducer("loms.order-events")
	if err != nil {
		myLogger.ErrorContext(context.Background(), "Failed to create kafka producer", zap.Error(err))
		return
	}
	serv := service.NewService(ordersRepo, stocksRepo, txManager, kafkaProducer)
	hand := handler.NewHandler(serv, tracer)

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		myLogger.ErrorContext(context.Background(), "Failed to listen", zap.Error(err))
		return
	}

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			mw.RestoreFromPanic,
			mw.Log,
			mw.Validate,
		),
		grpc.StatsHandler(otelgrpc.NewServerHandler(
			otelgrpc.WithTracerProvider(tp),
			otelgrpc.WithPropagators(otel.GetTextMapPropagator()),
		)),
	)

	grpcHealth := health.NewServer()
	healthpb.RegisterHealthServer(grpcServer, grpcHealth)
	grpcHealth.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	proto.RegisterLomsServiceServer(grpcServer, hand)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		myLogger.InfoContext(context.Background(), "Starting gRPC server")
		if err = grpcServer.Serve(lis); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			myLogger.ErrorContext(context.Background(), "Failed to start gRPC server", zap.Error(err))
		}
	}()

	// metrics
	metricsMux := http.NewServeMux()
	metricsMux.Handle("/metrics", metrics.MetricsHandler())

	go func() {
		slog.Info("Metrics server listening", "address", ":9091")
		if err := http.ListenAndServe(":9091", metricsMux); err != nil {
			slog.Error("Failed to serve metrics", "error", err)
		}
	}()

	<-stop
	myLogger.InfoContext(context.Background(), "Shutting down gRPC server")
	grpcServer.GracefulStop()
	myLogger.InfoContext(context.Background(), "Server gracefully stopped.")
}

func initDBPools() (*model.DBPools, error) {
	masterURL := os.Getenv("DB_MASTER_URL")
	replicaURL := os.Getenv("DB_REPLICA_URL")

	if masterURL == "" || replicaURL == "" {
		return nil, errors.New("database URLs must be set in DB_MASTER_URL and DB_REPLICA_URL")
	}
	masterCfg, err := pgxpool.ParseConfig(masterURL)
	masterPool, err := pgxpool.NewWithConfig(context.Background(), masterCfg)
	if err != nil {
		return nil, err
	}
	defer masterPool.Close()

	replicaCfg, err := pgxpool.ParseConfig(replicaURL)
	replicaPool, err := pgxpool.NewWithConfig(context.Background(), replicaCfg)
	if err != nil {
		return nil, err
	}
	defer masterPool.Close()

	return &model.DBPools{Master: masterPool, Replica: replicaPool}, nil
}
