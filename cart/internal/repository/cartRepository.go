package repository

import (
	"context"
	"fmt"
	"route256/cart/internal/model"
	"sync"
)

type CartRepository struct {
	carts map[int64]*model.DtoUserData
	mu    sync.RWMutex
}

func NewCartRepository() *CartRepository {
	return &CartRepository{
		carts: make(map[int64]*model.DtoUserData),
		mu:    sync.RWMutex{},
	}
}

func (r *CartRepository) AddItemToCart(ctx context.Context, userId int64, addItem *model.DtoItem) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	userData, ok := r.carts[userId]
	if !ok {
		userData = &model.DtoUserData{
			Items: map[int64]*model.DtoItem{
				addItem.SkuId: addItem,
			},
		}
	} else {
		gotItem, ok := userData.Items[addItem.SkuId]
		if !ok {
			userData.Items[addItem.SkuId] = addItem
		} else {
			gotItem.Count += addItem.Count
			userData.Items[addItem.SkuId] = gotItem
		}
	}

	r.carts[userId] = userData
	return nil
}

func (r *CartRepository) DeleteFromCart(ctx context.Context, userId int64, skuId int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	userData, ok := r.carts[userId]
	if !ok {
		return fmt.Errorf("Not found user %d", userId)
	}

	_, ok = userData.Items[skuId]
	if !ok {
		return fmt.Errorf("Not found product %d", skuId)
	}

	delete(r.carts[userId].Items, skuId)

	return nil
}

func (r *CartRepository) DeleteAllFromCart(ctx context.Context, userId int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.carts, userId)

	r.carts[userId] = &model.DtoUserData{
		Items: map[int64]*model.DtoItem{},
	}

	return nil
}

func (r *CartRepository) GetItems(ctx context.Context, userId int64) (*model.DtoUserData, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	userData, ok := r.carts[userId]
	if !ok {
		return &model.DtoUserData{
			Items: map[int64]*model.DtoItem{},
		}, nil
	}

	return userData, nil
}
