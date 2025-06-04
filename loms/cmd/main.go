package main

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"route256/loms/internal/handler"
	"route256/loms/internal/model"
	"route256/loms/internal/mw"
	"route256/loms/internal/repository"
	"route256/loms/internal/service"
	"route256/loms/proto"
	"syscall"
)

func main() {
	dbPool, err := initDBPools()
	if err != nil {
		slog.Error("Failed to connect to database")
		return
	}

	txManager := repository.NewTxManager(dbPool)
	ordersRepo := repository.NewOrdersRepository(dbPool)
	stocksRepo := repository.NewStocksRepository(dbPool)
	serv := service.NewService(ordersRepo, stocksRepo, txManager)
	hand := handler.NewHandler(serv)

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		slog.Error("failed to listen: %v", err)
		return
	}

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			mw.RestoreFromPanic,
			mw.Log,
			mw.Validate,
		),
	)
	proto.RegisterLomsServiceServer(grpcServer, hand)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		slog.Info("🚀 Mock LOMS gRPC Service running on :50051")
		if err := grpcServer.Serve(lis); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			slog.Error("Failed to serve gRPC server", "error", err)
		}
	}()

	<-stop
	slog.Info("Shutting down gRPC server...")
	grpcServer.GracefulStop()
	slog.Info("Server gracefully stopped.")
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
