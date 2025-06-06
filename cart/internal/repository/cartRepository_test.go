package repository

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"route256/cart/internal/model"
	"route256/cart/pkg/errgroup"
	"testing"
)

func TestAddItemToCart(t *testing.T) {
	// arrange
	var userId int64 = 111
	item := model.DtoItem{
		SkuId: 123,
		Count: 1,
	}
	ctx := context.Background()

	// act
	repo := NewCartRepository()
	err := repo.AddItemToCart(ctx, userId, &item)
	require.NoError(t, err)
	actual, err := repo.GetItems(ctx, userId)

	// assert
	assert.Equal(t, actual.Items[item.SkuId], &item)
}

func TestDeleteAllFromCart(t *testing.T) {
	// arrange
	var userId int64 = 111
	ctx := context.Background()
	item := model.DtoItem{
		SkuId: 123,
		Count: 1,
	}
	repo := NewCartRepository()
	err := repo.AddItemToCart(ctx, userId, &item)
	require.NoError(t, err)

	// act
	err = repo.DeleteAllFromCart(ctx, userId)
	require.NoError(t, err)
	actual, err := repo.GetItems(ctx, userId)

	// assert
	assert.Equal(t, len(actual.Items), 0)
}

func TestDeleteFromCart(t *testing.T) {
	// arrange
	var userId int64 = 111
	ctx := context.Background()
	item1 := model.DtoItem{
		SkuId: 1,
		Count: 1,
	}
	item2 := model.DtoItem{
		SkuId: 2,
		Count: 1,
	}
	repo := NewCartRepository()
	err := repo.AddItemToCart(ctx, userId, &item1)
	require.NoError(t, err)
	err = repo.AddItemToCart(ctx, userId, &item2)
	require.NoError(t, err)

	// act
	err = repo.DeleteFromCart(ctx, userId, item1.SkuId)
	require.NoError(t, err)
	actual, err := repo.GetItems(ctx, userId)

	// assert
	assert.Equal(t, actual.Items[item2.SkuId], &item2)
}

func TestGetItems_ReturnEmpty(t *testing.T) {
	// arrange
	var userId int64 = 111

	// act
	repo := NewCartRepository()
	actual, err := repo.GetItems(context.Background(), userId)
	require.NoError(t, err)

	// assert
	assert.Equal(t, len(actual.Items), 0)
}

func TestAddItemToCart_Concurrency(t *testing.T) {
	// arrange
	var userId int64 = 111
	var skuId int64 = 123
	totalItems := 100
	ctx := context.Background()
	g, ctx := errgroup.WithContext(ctx, totalItems)

	// act
	repo := NewCartRepository()
	for i := 0; i < totalItems; i++ {
		g.Go(func() error {
			item := model.DtoItem{
				SkuId: skuId,
				Count: 1,
			}
			return repo.AddItemToCart(ctx, userId, &item)
		})
	}

	if err := g.Wait(); err != nil {
		require.NoError(t, err)
	}

	actual, err := repo.GetItems(ctx, userId)
	require.NoError(t, err)

	// assert
	assert.Equal(t, actual.Items[skuId].SkuId, skuId)
	assert.Equal(t, actual.Items[skuId].Count, uint16(totalItems))
}
