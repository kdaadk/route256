package main

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"route256/cart/internal/handler"
	"route256/cart/internal/repository"
	"route256/cart/internal/service"
	"route256/cart/pkg/product"
)

func main() {
	productCli, err := product.NewClient()
	if err != nil {
		log.Fatalf("Init product provider failed: %v", err)
	}

	cartRepo := repository.NewCartRepository()
	serv := service.NewService(cartRepo, productCli)
	hand := handler.NewHandler(serv)
	router := mux.NewRouter()
	hand.RegisterRoutes(router)

	log.Println("🚀 Server running on :8082")
	if err := http.ListenAndServe("localhost:8082", router); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
