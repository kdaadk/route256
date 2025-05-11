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

func (r *CartRepository) AddItemToCart(userId int64, addItem *model.Item) error {
	userData, ok := r.carts[userId]
	if !ok {
		userData = &model.UserData{
			Items: map[int64]*model.Item{
				addItem.SkuId: addItem,
			},
			TotalPrice: addItem.Price * uint32(addItem.Count),
		}
	} else {
		userData.TotalPrice += addItem.Price
		gotItem, ok := userData.Items[addItem.SkuId]
		if !ok {
			userData.Items[addItem.SkuId] = addItem
		} else {
			gotItem.Price += addItem.Price
			gotItem.Count += addItem.Count

			userData.Items[addItem.SkuId] = gotItem
		}
	}

	r.carts[userId] = userData
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

	deletedTotalPrice := item.Price * uint32(item.Count)
	if deletedTotalPrice > userData.TotalPrice {
		return fmt.Errorf("Wrong total price, before: %d, subtract: %d", userData.TotalPrice, deletedTotalPrice)
	}

	delete(r.carts[userId].Items, skuId)
	r.carts[userId].TotalPrice = userData.TotalPrice - deletedTotalPrice

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
		return &model.UserData{
			Items:      map[int64]*model.Item{},
			TotalPrice: 0,
		}, nil
		//return &model.UserData{}, fmt.Errorf("No data for user %d", userId)
	}

	return userData, nil
}
