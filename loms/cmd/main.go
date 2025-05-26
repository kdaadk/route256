package main

import (
	"errors"
	"google.golang.org/grpc"
	"log"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"route256/loms/internal/handler"
	"route256/loms/internal/mw"
	"route256/loms/internal/repository"
	"route256/loms/internal/service"
	"route256/loms/proto"
	"syscall"
)

func main() {
	ordersRepo := repository.NewOrdersRepository()
	stocksRepo := repository.NewStocksRepository()
	serv := service.NewService(ordersRepo, stocksRepo)
	hand := handler.NewHandler(serv)

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
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
