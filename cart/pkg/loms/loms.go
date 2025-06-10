package loms

import (
	"context"
	"fmt"
	myLogger "github.com/kdaadk/route256/pkg/logger"
	"go.uber.org/zap"
	"route256/cart/internal/model"

	"route256/cart/vendor-proto/route256/loms"
)

func (c *Client) CreateOrder(ctx context.Context, userId int64, items []model.OrderItem) (int64, error) {
	req := &proto.CreateOrderRequest{
		UserId: userId,
		Items:  mapOrderItems(items),
	}

	res, err := c.client.CreateOrder(ctx, req)
	if err != nil {
		myLogger.ErrorContext(ctx, "Error creating order", zap.Error(err))
		return 0, fmt.Errorf("failed to create order: %w", err)
	}

	return res.OrderId, nil
}

func (c *Client) GetById(ctx context.Context, orderId int64) (*proto.GetByIdResponse, error) {
	req := &proto.GetByIdRequest{OrderId: orderId}
	res, err := c.client.GetOrderById(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get by id order: %w", err)
	}

	return res, nil
}

func (c *Client) PayOrder(ctx context.Context, orderId int64) error {
	req := &proto.PayOrderRequest{OrderId: orderId}
	_, err := c.client.PayOrder(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to pay order by id: %w", err)
	}

	return nil
}

func (c *Client) CancelOrder(ctx context.Context, orderId int64) error {
	req := &proto.CancelOrderRequest{OrderId: orderId}
	_, err := c.client.CancelOrder(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to cancel order by id: %w", err)
	}

	return nil
}

func (c *Client) GetStockInfos(ctx context.Context, skuIds []int64) (*proto.GetStockInfoResponse, error) {
	req := &proto.GetStockInfoRequest{SkuIds: skuIds}
	r, err := c.client.GetStockInfo(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get stock infos: %w", err)
	}

	return r, nil
}

func mapOrderItems(items []model.OrderItem) []*proto.OrderItem {
	res := make([]*proto.OrderItem, len(items))
	for i, item := range items {
		res[i] = &proto.OrderItem{
			Sku:   int32(item.SkuId),
			Count: int32(item.Count),
		}
	}

	return res
}
