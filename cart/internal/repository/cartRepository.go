package repository

import (
	"fmt"
	"route256/cart/internal/model"
)

type CartRepository struct {
	carts map[int64]*model.UserData
}

func NewCartRepository() *CartRepository {
	return &CartRepository{
		carts: make(map[int64]*model.UserData),
	}
}

func (r *CartRepository) AddItemToCart(userId int64, item *model.Item) error {
	if _, ok := r.carts[userId]; !ok {
		r.carts[userId] = &model.UserData{
			Items: map[int64]*model.Item{
				item.SkuId: item,
			},
			TotalPrice: item.Price,
		}
	}

	return nil
}

func (r *CartRepository) DeleteFromCart(userId int64, skuId int64) error {
	userData, ok := r.carts[userId]
	if !ok {
		return fmt.Errorf("Not found user %d", userId)
	}

	item, ok := userData.Items[skuId]
	if !ok {
		return fmt.Errorf("Not found product %d", skuId)
	}

	newTotalPrice := userData.TotalPrice - (item.Price * uint32(item.Count))
	if newTotalPrice < 0 {
		return fmt.Errorf("Wrong total price %d", newTotalPrice)
	}

	delete(r.carts[userId].Items, skuId)
	r.carts[userId].TotalPrice = newTotalPrice

	return nil
}

func (r *CartRepository) DeleteAllFromCart(userId int64) error {
	delete(r.carts, userId)

	r.carts[userId] = &model.UserData{
		Items:      map[int64]*model.Item{},
		TotalPrice: 0,
	}

	return nil
}

func (r *CartRepository) GetItems(userId int64) (*model.UserData, error) {
	userData, ok := r.carts[userId]
	if !ok {
		return &model.UserData{}, fmt.Errorf("No data for user %d", userId)
	}

	return userData, nil
}
