package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

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
)

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/product/{id}", handleGetProduct).Methods("GET")

	log.Println("🚀 Mock Product Service running on :8081")
	if err := http.ListenAndServe("localhost:8081", r); err != nil {
		log.Fatalf("mock product service failed: %v", err)
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
