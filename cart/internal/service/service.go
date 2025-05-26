package service

import (
	"errors"
	"fmt"
	"route256/cart/internal/model"
	"route256/cart/vendor-proto/route256/loms"
)

type productCli interface {
	GetProduct(productId int64) (*model.Product, error)
}

type cartRepo interface {
	AddItemToCart(userId int64, item *model.Item) error
	DeleteFromCart(userId int64, skuId int64) error
	DeleteAllFromCart(userId int64) error
	GetItems(userId int64) (*model.UserData, error)
}

type lomsCli interface {
	CreateOrder(userId int64, items []model.OrderItem) (int64, error)
	GetById(orderId int64) (*proto.GetByIdResponse, error)
	PayOrder(orderId int64) error
	CancelOrder(orderId int64) error
	GetStockInfos(skuIds []int64) (*proto.GetStockInfoResponse, error)
}

type Service struct {
	cartRepo   cartRepo
	productCli productCli
	lomsCli    lomsCli
}

func (s Service) AddItemToCart(userId, skuId int64, count uint32) error {
	prod, err := s.productCli.GetProduct(skuId)
	if err != nil {
		return err
	}

	item := model.Item{
		SkuId: skuId,
		Name:  prod.Name,
		Count: uint16(count),
		Price: prod.Price * count,
	}

	r, err := s.lomsCli.GetStockInfos([]int64{skuId})
	if err != nil {
		return err
	}

	if r.StockInfos[0].Count < count {
		return errors.New(fmt.Sprintf("not enough stock for sku %d", skuId))
	}

	return s.cartRepo.AddItemToCart(userId, &item)
}

func (s Service) DeleteFromCart(userId int64, skuId int64) error {
	return s.cartRepo.DeleteFromCart(userId, skuId)
}

func (s Service) GetItems(userId int64) (*model.UserData, error) {
	return s.cartRepo.GetItems(userId)
}

func (s Service) PayOrder(orderId int64) error {
	return s.lomsCli.PayOrder(orderId)
}

func (s Service) CancelOrder(orderId int64) error {
	return s.lomsCli.CancelOrder(orderId)
}

func (s Service) Checkout(userId int64) (int64, error) {
	userData, err := s.cartRepo.GetItems(userId)
	if err != nil {
		return 0, err
	}
	items := make([]model.OrderItem, 0)
	for _, item := range userData.Items {
		items = append(items, model.OrderItem{SkuId: item.SkuId, Count: item.Count})
	}

	orderId, err := s.lomsCli.CreateOrder(userId, items)
	if err != nil {
		return 0, err
	}

	err = s.cartRepo.DeleteAllFromCart(userId)
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
