package main

import (
	"context"
	"encoding/json"
	"errors"
	myLogger "github.com/kdaadk/route256/pkg/logger"
	"github.com/kdaadk/route256/pkg/tracing"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"net/http"
	"os"
	"os/signal"
	"route256/product/internal/mw"
	"strconv"
	"syscall"
	"time"

	"github.com/gorilla/mux"
)

type Product struct {
	SkuId int64  `json:"sku_id"`
	Name  string `json:"name"`
	Count uint16 `json:"count"`
	Price uint32 `json:"price"`
}

var (
	storage = map[int64]Product{
		1076963: {SkuId: 1076963, Name: "Item #1076963", Count: 6, Price: 100},
		1148162: {SkuId: 1148162, Name: "Item #1148162", Count: 16, Price: 200},
	}
	serviceName = "product-service"
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
	tp, err := tracing.InitTracer("jaeger:4317", serviceName, tracing.GRPCTransport)
	if err != nil {
		myLogger.ErrorContext(context.Background(), "Failed to initialize tracer", zap.Error(err))
		return
	}
	defer tracing.ShutdownTracer(tp)

	// server
	r := mux.NewRouter()
	r.HandleFunc("/product/{id}", handleGetProduct).Methods("GET")
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrappedRouter := mw.TracingMiddleware(r)

	server := &http.Server{
		Addr:         "0.0.0.0:8081",
		Handler:      wrappedRouter,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
		IdleTimeout:  2 * time.Second,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		myLogger.InfoContext(context.Background(), "🚀 Server running on :8081")
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			myLogger.ErrorContext(context.Background(), "Server failed on :8081", zap.Error(err))
		}
	}()

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

func handleGetProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr, ok := vars["id"]
	if !ok {
		http.Error(w, "missing product Id", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid product Id", http.StatusBadRequest)
		return
	}

	product, ok := storage[id]
	if !ok {
		http.Error(w, "invalid sku", http.StatusPreconditionFailed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(product); err != nil {
		http.Error(w, "failed to encode product", http.StatusInternalServerError)
	}
}
