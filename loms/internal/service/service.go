package service

import (
	"errors"
	"route256/loms/internal/model"
	proto "route256/loms/proto"
)

type ordersRepo interface {
	CreateOrder(req *proto.CreateOrderRequest) (int64, error)
	SetState(orderId int64, status model.OrderStatus) error
	GetById(orderId int64) (*model.Order, error)
}

type stocksRepo interface {
	Reserve(skuId int64, count uint32) (model.OrderStatus, error)
	CancelReserve(skuId int64, count uint32) error
	RemoveReserve(skuId int64, count uint32) error
	GetStockInfos(skuIds []int64) (*[]model.Stock, error)
}

type Service struct {
	ordersRepo ordersRepo
	stocksRepo stocksRepo
}

func NewService(ordersRepo ordersRepo, stocksRepo stocksRepo) *Service {
	return &Service{
		ordersRepo: ordersRepo,
		stocksRepo: stocksRepo,
	}
}

func (s Service) CreateOrder(req *proto.CreateOrderRequest) (int64, error) {
	orderId, err := s.ordersRepo.CreateOrder(req)
	if err != nil {
		return 0, err
	}

	status, err := s.reserve(req)
	if err != nil && status == model.OrderStatus_Failed {
		return 0, err
	}

	err = s.ordersRepo.SetState(orderId, status)
	if err != nil {
		return 0, err
	}

	return orderId, nil
}

func (s Service) GetById(orderId int64) (*model.Order, error) {
	return s.ordersRepo.GetById(orderId)
}

func (s Service) reserve(req *proto.CreateOrderRequest) (model.OrderStatus, error) {
	reserved := map[int64]uint32{}
	failed := false

	for _, item := range req.Items {
		status, err := s.stocksRepo.Reserve(int64(item.Sku), uint32(item.Count))
		if err != nil && status == model.OrderStatus_Failed {
			failed = true
		} else {
			reserved[int64(item.Sku)] = uint32(item.Count)
		}
	}

	if failed {
		for skuId, count := range reserved {
			err := s.stocksRepo.CancelReserve(skuId, count)
			if err != nil {
				panic(err)
			}
		}

		return model.OrderStatus_Failed, errors.New("failed to reserve")
	} else {
		return model.OrderStatus_AwaitingPayment, nil
	}
}

func (s Service) PayOrder(req *proto.PayOrderRequest) error {
	order, err := s.ordersRepo.GetById(req.OrderId)
	if err != nil {
		return err
	}

	for _, item := range order.Items {
		err = s.stocksRepo.RemoveReserve(item.SkuId, item.Count)
		if err != nil {
			return err
		}
	}

	err = s.ordersRepo.SetState(order.OrderId, model.OrderStatus_Paid)
	if err != nil {
		return err
	}

	return nil
}

func (s Service) CancelOrder(req *proto.CancelOrderRequest) error {
	order, err := s.ordersRepo.GetById(req.OrderId)
	if err != nil {
		return err
	}

	for _, item := range order.Items {
		err = s.stocksRepo.CancelReserve(item.SkuId, item.Count)
		if err != nil {
			return err
		}
	}

	err = s.ordersRepo.SetState(order.OrderId, model.OrderStatus_Cancelled)
	if err != nil {
		return err
	}

	return nil
}

func (s Service) GetStockInfos(req *proto.GetStockInfoRequest) (*[]model.Stock, error) {
	return s.stocksRepo.GetStockInfos(req.SkuIds)
}
