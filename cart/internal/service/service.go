package service

import (
	"context"
	"errors"
	"fmt"
	"route256/cart/internal/model"
	"route256/cart/pkg/errgroup"
	"route256/cart/vendor-proto/route256/loms"
	"sync"
)

type productCli interface {
	GetProduct(ctx context.Context, productId int64) (*model.Product, error)
}

type cartRepo interface {
	AddItemToCart(ctx context.Context, userId int64, item *model.DtoItem) error
	DeleteFromCart(ctx context.Context, userId int64, skuId int64) error
	DeleteAllFromCart(ctx context.Context, userId int64) error
	GetItems(ctx context.Context, userId int64) (*model.DtoUserData, error)
}

type lomsCli interface {
	CreateOrder(ctx context.Context, userId int64, items []model.OrderItem) (int64, error)
	GetById(ctx context.Context, orderId int64) (*proto.GetByIdResponse, error)
	PayOrder(ctx context.Context, orderId int64) error
	CancelOrder(ctx context.Context, orderId int64) error
	GetStockInfos(ctx context.Context, skuIds []int64) (*proto.GetStockInfoResponse, error)
}

type Service struct {
	cartRepo   cartRepo
	productCli productCli
	lomsCli    lomsCli
}

func (s Service) AddItemToCart(ctx context.Context, userId, skuId int64, count uint32) error {
	prod, err := s.productCli.GetProduct(ctx, skuId)
	if err != nil {
		return err
	}

	if prod == nil {
		return errors.New("product not found")
	}

	item := model.DtoItem{
		SkuId: skuId,
		Count: uint16(count),
	}

	r, err := s.lomsCli.GetStockInfos(ctx, []int64{skuId})
	if err != nil {
		return err
	}

	if r.StockInfos[0].Count < count {
		return errors.New(fmt.Sprintf("not enough stock for sku %d", skuId))
	}

	return s.cartRepo.AddItemToCart(ctx, userId, &item)
}

func (s Service) DeleteFromCart(ctx context.Context, userId int64, skuId int64) error {
	return s.cartRepo.DeleteFromCart(ctx, userId, skuId)
}

func (s Service) GetItems(ctx context.Context, userId int64) (*model.UserData, error) {
	dtoUserData, err := s.cartRepo.GetItems(ctx, userId)
	if err != nil {
		return nil, err
	}

	userData := &model.UserData{Items: map[int64]*model.Item{}}
	var totalPrice uint32 = 0

	g, ctx := errgroup.WithContext(ctx, 10)
	mu := &sync.Mutex{}

	for _, item := range dtoUserData.Items {
		g.Go(func() error {
			prod, err := s.productCli.GetProduct(ctx, item.SkuId)
			if err != nil {
				return err
			}

			if prod == nil {
				return errors.New("product not found")
			}

			mu.Lock()
			userData.Items[item.SkuId] = &model.Item{
				SkuId: item.SkuId,
				Count: item.Count,
				Price: prod.Price,
				Name:  prod.Name,
			}
			mu.Unlock()

			return nil
		})
	}

	if err = g.Wait(); err != nil {
		return nil, err
	}

	userData.TotalPrice = totalPrice
	return userData, nil
}

func (s Service) PayOrder(ctx context.Context, orderId int64) error {
	return s.lomsCli.PayOrder(ctx, orderId)
}

func (s Service) CancelOrder(ctx context.Context, orderId int64) error {
	return s.lomsCli.CancelOrder(ctx, orderId)
}

func (s Service) Checkout(ctx context.Context, userId int64) (int64, error) {
	userData, err := s.cartRepo.GetItems(ctx, userId)
	if err != nil {
		return 0, err
	}
	items := make([]model.OrderItem, 0)
	for _, item := range userData.Items {
		items = append(items, model.OrderItem{SkuId: item.SkuId, Count: item.Count})
	}

	orderId, err := s.lomsCli.CreateOrder(ctx, userId, items)
	if err != nil {
		return 0, err
	}

	err = s.cartRepo.DeleteAllFromCart(ctx, userId)
	if err != nil {
		return 0, err
	}

	return orderId, nil
}

func NewService(cartRepo cartRepo, productCli productCli, lomsCli lomsCli) *Service {
	return &Service{
		cartRepo:   cartRepo,
		productCli: productCli,
		lomsCli:    lomsCli,
	}
}
