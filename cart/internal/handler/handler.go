package handler

import (
	"context"
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
	AddItemToCart(ctx context.Context, userId, skuId int64, count uint32) error
	DeleteFromCart(ctx context.Context, userId int64, skuId int64) error
	GetItems(ctx context.Context, userId int64) (*model.UserData, error)

	PayOrder(ctx context.Context, orderId int64) error
	CancelOrder(ctx context.Context, orderId int64) error

	Checkout(ctx context.Context, userId int64) (int64, error)
}

type Handler struct {
	service cartService
}

func NewHandler(service cartService) *Handler {
	return &Handler{
		service: service}
}

var (
	AddToCartRoute      = "/user/{userId}/cart/{skuId}"
	DeleteFromCartRoute = "/user/{userId}/cart/{skuId}"
	GetAllFromCartRoute = "/user/{userId}/cart"

	PayOrderRoute    = "/order/pay/{orderId}"
	CancelOrderRoute = "/order/cancel/{orderId}"

	CheckoutRoute = "/cart/checkout"
)

func (h *Handler) RegisterRoutes(r *mux.Router) {
	// cart/item/add - добавляем в корзину и проверяем, что есть в наличии
	r.HandleFunc(AddToCartRoute, h.AddToCartHandler).Methods("POST")
	// cart/item/delete - можем удалять из корзины
	r.HandleFunc(DeleteFromCartRoute, h.DeleteFromCartHandler).Methods("DELETE")
	// cart/list - можем получать список товаров корзины
	r.HandleFunc(GetAllFromCartRoute, h.GetAllFromCartHandler).Methods("GET")

	// order/pay - оплачиваем заказ
	r.HandleFunc(PayOrderRoute, h.PayOrderHandler).Methods("POST")
	// order/cancel - отмена заказа до оплаты
	r.HandleFunc(CancelOrderRoute, h.CancelOrderHandler).Methods("POST")

	// cart/checkout - приобретаем товары через Checkout
	r.HandleFunc(CheckoutRoute, h.CheckoutHandler).Methods("POST")
}

func (h *Handler) DeleteFromCartHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
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

	if err = h.service.DeleteFromCart(ctx, userId, skuId); err != nil {
		writeErr(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) GetAllFromCartHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	rawUserId := vars["userId"]

	userId, err := strconv.ParseInt(rawUserId, 10, 64)
	if err != nil || userId <= 0 {
		writeErr(w, err)
		return
	}

	userData, err := h.service.GetItems(ctx, userId)
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
	ctx := r.Context()
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

	err = h.service.AddItemToCart(ctx, userId, skuId, uint32(req.Count))
	if err != nil {
		writeErr(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) PayOrderHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	rawOrderId := vars["orderId"]

	orderId, err := strconv.ParseInt(rawOrderId, 10, 64)
	if err != nil || orderId <= 0 {
		writeErr(w, err)
		return
	}

	err = h.service.PayOrder(ctx, orderId)
	if err != nil {
		writeErr(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) CancelOrderHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	rawOrderId := vars["orderId"]

	orderId, err := strconv.ParseInt(rawOrderId, 10, 64)
	if err != nil || orderId <= 0 {
		writeErr(w, err)
		return
	}

	err = h.service.CancelOrder(ctx, orderId)
	if err != nil {
		writeErr(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) CheckoutHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	rawUserId := vars["user"]

	userId, err := strconv.ParseInt(rawUserId, 10, 64)
	if err != nil || userId <= 0 {
		writeErr(w, err)
		return
	}

	orderId, err := h.service.Checkout(ctx, userId)
	if err != nil {
		writeErr(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(orderId)
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
