package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"route256/cart/internal/model"
	"strconv"
)

type cartService interface {
	AddItemToCart(userId int64, item *model.Item) error
	DeleteFromCart(userId int64, skuId int64) error
	DeleteAllFromCart(userId int64) error
	GetItems(userId int64) (*model.UserData, error)
	GetProduct(productId int64) (*model.Product, error)
}

type Handler struct {
	service cartService
}

func NewHandler(service cartService) *Handler {
	return &Handler{
		service: service}
}

var (
	AddToCartRoute         = "/user/{userId}/cart/{skuId}"
	DeleteFromCartRoute    = "/user/{userId}/cart/{skuId}"
	DeleteAllFromCartRoute = "/user/{userId}/cart"
	GetAllFromCartRoute    = "/user/{userId}/cart"
)

func (h *Handler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc(AddToCartRoute, h.AddToCartHandler).Methods("POST")
	r.HandleFunc(DeleteFromCartRoute, h.DeleteFromCartHandler).Methods("DELETE")
	r.HandleFunc(DeleteAllFromCartRoute, h.DeleteAllFromCartHandler).Methods("DELETE")
	r.HandleFunc(GetAllFromCartRoute, h.GetAllFromCartHandler).Methods("GET")
}

func (h *Handler) DeleteFromCartHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	rawUserId := vars["userId"]
	rawSkuId := vars["skuId"]

	userId, err := strconv.ParseInt(rawUserId, 10, 64)
	if err != nil || userId <= 0 {
		writeErr(w, err)
		return
	}

	skuId, err := strconv.ParseInt(rawSkuId, 10, 64)
	if err != nil || skuId <= 0 {
		writeErr(w, err)
		return
	}

	if err = h.service.DeleteFromCart(userId, skuId); err != nil {
		writeErr(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) DeleteAllFromCartHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	rawUserId := vars["userId"]

	userId, err := strconv.ParseInt(rawUserId, 10, 64)
	if err != nil || userId <= 0 {
		writeErr(w, err)
		return
	}

	err = h.service.DeleteAllFromCart(userId)
	if err != nil {
		writeErr(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) GetAllFromCartHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	rawUserId := vars["userId"]

	userId, err := strconv.ParseInt(rawUserId, 10, 64)
	if err != nil || userId <= 0 {
		writeErr(w, err)
		return
	}

	userData, err := h.service.GetItems(userId)
	if err != nil {
		writeErr(w, err)
		return
	}

	items := make([]model.Item, 0)
	for _, item := range userData.Items {
		items = append(items, *item)
	}

	response := model.GetItemsFromCartResponse{
		TotalPrice: userData.TotalPrice,
		Items:      items,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		return
	}
}

func (h *Handler) AddToCartHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	rawUserId := vars["userId"]
	rawSkuId := vars["skuId"]

	userId, err := strconv.ParseInt(rawUserId, 10, 64)
	if err != nil || userId <= 0 {
		writeErr(w, err)
		return
	}
	skuId, err := strconv.ParseInt(rawSkuId, 10, 64)
	if err != nil || skuId <= 0 {
		writeErr(w, err)
		return
	}

	var req model.AddToCartRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if req.Count == 0 {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	prod, err := h.service.GetProduct(skuId)
	if err != nil {
		writeErr(w, err)
		return
	}

	item := model.Item{
		SkuId: skuId,
		Name:  prod.Name,
		Count: req.Count,
		Price: prod.Price * uint32(req.Count),
	}

	err = h.service.AddItemToCart(userId, &item)
	if err != nil {
		writeErr(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func writeErr(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")

	log.Printf("line 169, error: %v", err)
	if errors.Is(err, model.ErrPreconditionFailed) {
		w.WriteHeader(http.StatusPreconditionFailed)
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}

	_, errOut := fmt.Fprintf(w, "{\"message\":\"%s\"}", err)
	if errOut != nil {
		log.Printf("failed %s", errOut.Error())
		return
	}
}
