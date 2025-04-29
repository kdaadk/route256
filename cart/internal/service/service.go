package service

import (
	"route256/cart/internal/model"
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

type Service struct {
	cartRepo   cartRepo
	productCli productCli
}

func (s Service) AddItemToCart(userId int64, item *model.Item) error {
	return s.cartRepo.AddItemToCart(userId, item)
}

func (s Service) DeleteFromCart(userId int64, skuId int64) error {
	return s.cartRepo.DeleteFromCart(userId, skuId)
}

func (s Service) DeleteAllFromCart(userId int64) error {
	return s.cartRepo.DeleteAllFromCart(userId)
}

func (s Service) GetItems(userId int64) (*model.UserData, error) {
	return s.cartRepo.GetItems(userId)
}

func (s Service) GetProduct(productId int64) (*model.Product, error) {
	return s.productCli.GetProduct(productId)
}

func NewService(cartRepo cartRepo, productCli productCli) *Service {
	return &Service{
		cartRepo:   cartRepo,
		productCli: productCli,
	}
}
