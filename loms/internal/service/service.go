package service

import (
	"context"
	"github.com/jackc/pgx/v5"
	"route256/loms/internal/model"
	proto "route256/loms/proto"
)

type ordersRepo interface {
	CreateOrder(tx pgx.Tx, userId int64, items []*model.Item) (int64, error)
	SetState(tx pgx.Tx, orderId int64, status model.OrderStatus) error
	GetById(orderId int64) (*model.Order, error)
}

type stocksRepo interface {
	Reserve(tx pgx.Tx, items []*model.Item) error
	CancelReserve(tx pgx.Tx, skuId int64, count uint32) error
	RemoveReserve(tx pgx.Tx, skuId int64, count uint32) error
	GetStockInfos(skuIds []int64) ([]*model.Stock, error)
}

type TxManager interface {
	Tx(ctx context.Context, fn func(ctx context.Context, tx pgx.Tx) error, opts *pgx.TxOptions) error
}

type Service struct {
	ordersRepo ordersRepo
	stocksRepo stocksRepo
	txManager  TxManager
}

func NewService(ordersRepo ordersRepo, stocksRepo stocksRepo, txManager TxManager) *Service {
	return &Service{
		ordersRepo: ordersRepo,
		stocksRepo: stocksRepo,
		txManager:  txManager,
	}
}

func (s Service) CreateOrder(req *proto.CreateOrderRequest) (int64, error) {
	items := make([]*model.Item, 0)
	for _, item := range req.Items {
		items = append(items, &model.Item{SkuId: int64(item.Sku), Count: uint32(item.Count)})
	}

	var orderId int64
	err := s.txManager.Tx(context.Background(), func(ctx context.Context, tx pgx.Tx) error {
		var err error
		orderId, err = s.ordersRepo.CreateOrder(tx, req.UserId, items)
		if err != nil {
			return err
		}

		err = s.stocksRepo.Reserve(tx, items)
		if err != nil {
			return err
		}

		err = s.ordersRepo.SetState(tx, orderId, model.OrderStatus_AwaitingPayment)
		if err != nil {
			return err
		}

		return nil
	}, nil)

	if err != nil {
		return 0, err
	}

	return orderId, nil
}

func (s Service) GetById(orderId int64) (*model.Order, error) {
	return s.ordersRepo.GetById(orderId)
}

func (s Service) PayOrder(req *proto.PayOrderRequest) error {
	err := s.txManager.Tx(context.Background(), func(ctx context.Context, tx pgx.Tx) error {
		order, err := s.ordersRepo.GetById(req.OrderId)
		if err != nil {
			return err
		}

		for _, item := range order.Items {
			err = s.stocksRepo.RemoveReserve(tx, item.SkuId, item.Count)
			if err != nil {
				return err
			}
		}

		err = s.ordersRepo.SetState(tx, order.OrderId, model.OrderStatus_Paid)
		if err != nil {
			return err
		}

		return nil
	}, nil)

	if err != nil {
		return err
	}

	return nil
}

func (s Service) CancelOrder(req *proto.CancelOrderRequest) error {
	err := s.txManager.Tx(context.Background(), func(ctx context.Context, tx pgx.Tx) error {
		order, err := s.ordersRepo.GetById(req.OrderId)
		if err != nil {
			return err
		}

		for _, item := range order.Items {
			err = s.stocksRepo.CancelReserve(tx, item.SkuId, item.Count)
			if err != nil {
				return err
			}
		}

		err = s.ordersRepo.SetState(tx, order.OrderId, model.OrderStatus_Cancelled)
		if err != nil {
			return err
		}

		return nil
	}, nil)

	if err != nil {
		return err
	}

	return nil
}

func (s Service) GetStockInfos(req *proto.GetStockInfoRequest) ([]*model.Stock, error) {
	return s.stocksRepo.GetStockInfos(req.SkuIds)
}
