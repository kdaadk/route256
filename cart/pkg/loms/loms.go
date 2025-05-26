package loms

import (
	"context"
	"fmt"
	"google.golang.org/grpc/metadata"
	"log/slog"
	"route256/cart/internal/model"

	"route256/cart/vendor-proto/route256/loms"
)

func (c *Client) CreateOrder(userId int64, items []model.OrderItem) (int64, error) {
	req := &proto.CreateOrderRequest{
		UserId: userId,
		Items:  mapOrderItems(items),
	}

	md := metadata.New(map[string]string{
		"x-request-id": "test",
	})
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	res, err := c.client.CreateOrder(ctx, req)
	if err != nil {
		slog.Error(err.Error())
		return 0, fmt.Errorf("failed to create order: %w", err)
	}

	return res.OrderId, nil
}

func (c *Client) GetById(orderId int64) (*proto.GetByIdResponse, error) {
	req := &proto.GetByIdRequest{OrderId: orderId}

	md := metadata.New(map[string]string{
		"x-request-id": "test",
	})
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	res, err := c.client.GetOrderById(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get by id order: %w", err)
	}

	return res, nil
}

func (c *Client) PayOrder(orderId int64) error {
	req := &proto.PayOrderRequest{OrderId: orderId}

	md := metadata.New(map[string]string{
		"x-request-id": "test",
	})
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	_, err := c.client.PayOrder(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to pay order by id: %w", err)
	}

	return nil
}

func (c *Client) CancelOrder(orderId int64) error {
	req := &proto.CancelOrderRequest{OrderId: orderId}

	md := metadata.New(map[string]string{
		"x-request-id": "test",
	})
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	_, err := c.client.CancelOrder(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to cancel order by id: %w", err)
	}

	return nil
}

func (c *Client) GetStockInfos(skuIds []int64) (*proto.GetStockInfoResponse, error) {
	req := &proto.GetStockInfoRequest{SkuIds: skuIds}

	md := metadata.New(map[string]string{
		"x-request-id": "test",
	})
	ctx := metadata.NewOutgoingContext(context.Background(), md)
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
