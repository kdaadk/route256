package main

import (
	"context"
	"errors"
	"github.com/gorilla/mux"
	myLogger "github.com/kdaadk/route256/pkg/logger"
	"github.com/kdaadk/route256/pkg/metrics"
	"github.com/kdaadk/route256/pkg/tracing"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
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

var serviceName = "cart-service"

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
	tp, err := tracing.InitTracer("jaeger:4317", serviceName, tracing.GRPCTransport)
	if err != nil {
		myLogger.ErrorContext(context.Background(), "Failed to initialize tracer", zap.Error(err))
		return
	}
	defer tracing.ShutdownTracer(tp)

	tracer := tp.Tracer(serviceName)

	// product client
	productCli, err := product.NewClient("http://product:8081", tp)
	if err != nil {
		myLogger.ErrorContext(context.Background(), "Failed to create product client", zap.Error(err))
		return
	}
	myLogger.InfoContext(context.Background(), "🚀 Server productCli running on :8081")

	// loms client
	lomsCli, _ := loms.NewClient("loms:50051")
	err = lomsCli.Connect()
	if err != nil {
		myLogger.ErrorContext(context.Background(), "Failed to connect loms client", zap.Error(err))
		return
	}
	myLogger.InfoContext(context.Background(), "🚀 Server lomsCli running on :50051")

	// server
	cartRepo := repository.NewCartRepository()
	serv := service.NewService(cartRepo, productCli, lomsCli, tracer)
	hand := handler.NewHandler(serv, tracer)
	router := mux.NewRouter()
	hand.RegisterRoutes(router)

	wrappedRouter := mw.TracingMiddleware(router)

	server := &http.Server{
		Addr:         ":8082",
		Handler:      wrappedRouter,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
		IdleTimeout:  2 * time.Second,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	myLogger.InfoContext(context.Background(), "🟢 Attempting to start HTTP server on :8082")

	go func() {
		myLogger.InfoContext(context.Background(), "🚀 Server running on :8082")
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			myLogger.ErrorContext(context.Background(), "Server failed on :8082", zap.Error(err))
		}
	}()

	// metrics
	metricsMux := http.NewServeMux()
	metricsMux.Handle("/metrics", metrics.MetricsHandler())
	go func() {
		myLogger.InfoContext(context.Background(), "Metrics server listening 9092")
		if err := http.ListenAndServe(":9093", metricsMux); err != nil {
			myLogger.ErrorContext(context.Background(), "Failed to serve metrics", zap.Error(err))
		}
	}()

	// waiting for gracefully shutdown
	<-quit
	myLogger.InfoContext(context.Background(), "Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		myLogger.ErrorContext(ctx, "Server shutdown failed", zap.Error(err))
	} else {
		myLogger.InfoContext(context.Background(), "Server exited gracefully")
	}
}
