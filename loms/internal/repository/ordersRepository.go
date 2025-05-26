package repository

import (
	"errors"
	"math/rand"
	"route256/loms/internal/model"
	proto "route256/loms/proto"
)

type OrdersRepository struct {
	orders map[int64]*model.Order
}

func NewOrdersRepository() *OrdersRepository {
	return &OrdersRepository{
		orders: make(map[int64]*model.Order),
	}
}

func (r *OrdersRepository) CreateOrder(req *proto.CreateOrderRequest) (int64, error) {
	orderId := rand.Int63n(100000000)
	items := make([]model.Item, 0, len(req.Items))
	for _, item := range req.Items {
		items = append(items, model.Item{SkuId: int64(item.Sku), Count: uint32(item.Count)})
	}

	order := &model.Order{OrderId: orderId, UserId: req.UserId, Items: items, Status: model.OrderStatus_New}
	r.orders[order.OrderId] = order
	return order.OrderId, nil
}

func (r *OrdersRepository) SetState(orderId int64, status model.OrderStatus) error {
	order, ok := r.orders[orderId]
	if !ok {
		return errors.New("order not found")
	}

	order.Status = status

	return nil
}

func (r *OrdersRepository) GetById(orderId int64) (*model.Order, error) {
	order, ok := r.orders[orderId]
	if !ok {
		return nil, errors.New("order not found")
	}

	return order, nil
}
