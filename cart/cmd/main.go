package main

import (
	"context"
	"errors"
	"github.com/gorilla/mux"
	myLogger "github.com/kdaadk/route256/pkg/logger"
	"github.com/kdaadk/route256/pkg/tracing"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log"
	"net/http"
	"os"
	"os/signal"
	"route256/cart/internal/handler"
	"route256/cart/internal/mw"
	"route256/cart/internal/repository"
	"route256/cart/internal/service"
	"route256/cart/pkg/loms"
	"route256/cart/pkg/product"
	"syscall"
	"time"
)

func main() {
	// logging
	config := zap.NewDevelopmentConfig()
	config.ErrorOutputPaths = []string{"output"}
	config.Level.SetLevel(zapcore.InfoLevel)
	logger := myLogger.NewMyLogger(config)
	defer func(logger *myLogger.MyLogger) {
		err := logger.Sync()
		if err != nil {
			panic(err)
		}
	}(logger)

	// tracing
	tp, err := initTracer()
	defer tracing.ShutdownTracer(tp)

	// product client
	productCli, err := product.NewClient("http://localhost:8081")
	if err != nil {
		myLogger.ErrorContext(context.Background(), "Failed to create product client", zap.Error(err))
		return
	}
	log.Println("🚀 Server productCli running on :8081")

	// loms client
	lomsCli, _ := loms.NewClient("localhost:50051")
	err = lomsCli.Run()
	if err != nil {
		myLogger.ErrorContext(context.Background(), "Failed to run loms client", zap.Error(err))
		return
	}
	myLogger.InfoContext(context.Background(), "🚀 Server lomsCli running on :50051")

	// server
	cartRepo := repository.NewCartRepository()
	serv := service.NewService(cartRepo, productCli, lomsCli)
	hand := handler.NewHandler(serv)
	router := mux.NewRouter()
	hand.RegisterRoutes(router)
	tracedRouter := mw.TracingMiddleware(router)

	server := &http.Server{
		Addr:         ":8082",
		Handler:      tracedRouter,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
		IdleTimeout:  2 * time.Second,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		myLogger.InfoContext(context.Background(), "🚀 Server running on :8082")
		if err = server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			myLogger.ErrorContext(context.Background(), "Server failed on :8082", zap.Error(err))
		}
	}()

	<-quit
	myLogger.InfoContext(context.Background(), "Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err = server.Shutdown(ctx); err != nil {
		myLogger.ErrorContext(ctx, "Server shutdown failed", zap.Error(err))
	} else {
		myLogger.InfoContext(context.Background(), "Server exited gracefully")
	}
}

func initTracer() (*sdktrace.TracerProvider, error) {
	host := "localhost"
	port := "5432"
	serviceName := "loms"
	tp, err := tracing.InitTracer(host+":"+port, serviceName, tracing.GRPCTransport)
	if err != nil {
		return nil, err
	}
	return tp, nil
}
