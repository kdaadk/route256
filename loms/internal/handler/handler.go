package handler

import (
	"context"
	"errors"
	myLogger "github.com/kdaadk/route256/pkg/logger"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/types/known/emptypb"
	"route256/loms/internal/mappers"
	"route256/loms/internal/service"
	"route256/loms/proto"
)

var _ proto.LomsServiceServer = (*Handler)(nil)

type Handler struct {
	Service *service.Service
	tracer  trace.Tracer
	proto.UnimplementedLomsServiceServer
}

func NewHandler(service *service.Service, tracer trace.Tracer) *Handler {
	return &Handler{Service: service, tracer: tracer}
}

func (h *Handler) CreateOrder(ctx context.Context, req *proto.CreateOrderRequest) (*proto.CreateOrderResponse, error) {
	ctx, span := h.tracer.Start(ctx, "CreateOrder")
	defer span.End()
	myLogger.InfoContext(ctx, "create_order")

	orderId, err := h.Service.CreateOrder(req)
	if err != nil {
		return nil, err
	}

	response := proto.CreateOrderResponse{OrderId: orderId}
	return &response, nil
}

func (h *Handler) GetOrderById(ctx context.Context, req *proto.GetByIdRequest) (*proto.GetByIdResponse, error) {
	ctx, span := h.tracer.Start(ctx, "GetOrderById")
	defer span.End()

	order, err := h.Service.GetById(req.OrderId)
	if order == nil || err != nil {
		return nil, errors.New("order not found")
	}

	items := make([]*proto.OrderItem, 0)
	for _, item := range order.Items {
		items = append(items, &proto.OrderItem{Sku: int32(item.SkuId), Count: int32(item.Count)})
	}
	response := proto.GetByIdResponse{Status: mappers.MapOrderStatusToBll(order.Status), UserId: order.UserId, Items: items}

	return &response, nil
}

func (h *Handler) PayOrder(ctx context.Context, req *proto.PayOrderRequest) (*emptypb.Empty, error) {
	ctx, span := h.tracer.Start(ctx, "PayOrder")
	defer span.End()

	err := h.Service.PayOrder(req)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (h *Handler) CancelOrder(ctx context.Context, req *proto.CancelOrderRequest) (*emptypb.Empty, error) {
	ctx, span := h.tracer.Start(ctx, "CancelOrder")
	defer span.End()

	err := h.Service.CancelOrder(req)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (h *Handler) GetStockInfo(ctx context.Context, req *proto.GetStockInfoRequest) (*proto.GetStockInfoResponse, error) {
	ctx, span := h.tracer.Start(ctx, "GetStockInfo")
	defer span.End()

	stocks, err := h.Service.GetStockInfos(req)
	if err != nil {
		return nil, err
	}

	stockInfos := make([]*proto.StockInfo, 0)
	for _, s := range stocks {
		stockInfos = append(stockInfos, &proto.StockInfo{SkuId: s.SkuId, Count: s.TotalCount - s.Reserved})
	}
	return &proto.GetStockInfoResponse{StockInfos: stockInfos}, nil
}
