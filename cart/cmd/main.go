package main

import (
	"context"
	"errors"
	"github.com/gorilla/mux"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"route256/cart/internal/handler"
	"route256/cart/internal/repository"
	"route256/cart/internal/service"
	"route256/cart/pkg/loms"
	"route256/cart/pkg/product"
	"syscall"
	"time"
)

func main() {
	productCli, err := product.NewClient("http://localhost:8081")
	if err != nil {
		log.Fatalf("Init product cli failed: %v", err)
	}
	log.Println("🚀 Server productCli running on :8081")

	lomsCli, _ := loms.NewClient("localhost:50051")
	err = lomsCli.Run()
	if err != nil {
		slog.Error("Failed to start loms client", "error", err)
		return
	}
	log.Println("🚀 Server lomsCli running on :50051")

	cartRepo := repository.NewCartRepository()
	serv := service.NewService(cartRepo, productCli, lomsCli)
	hand := handler.NewHandler(serv)
	router := mux.NewRouter()
	hand.RegisterRoutes(router)

	server := &http.Server{
		Addr:         ":8082",
		Handler:      router,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
		IdleTimeout:  2 * time.Second,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		slog.Info("🚀 Server running on :8082")
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("Server failed", "error", err)
		}
	}()

	<-quit
	slog.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
	} else {
		slog.Info("Server exited gracefully")
	}
}
