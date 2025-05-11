package repository

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"route256/cart/internal/model"
	"testing"
)

func TestAddItemToCart(t *testing.T) {
	// arrange
	var userId int64 = 111
	item := model.Item{
		SkuId: 123,
		Name:  "test",
		Count: 1,
		Price: 100,
	}

	// act
	repo := NewCartRepository()
	err := repo.AddItemToCart(userId, &item)
	require.NoError(t, err)
	actual, err := repo.GetItems(userId)

	// assert
	assert.Equal(t, actual.Items[item.SkuId], &item)
}

func TestDeleteAllFromCart(t *testing.T) {
	// arrange
	var userId int64 = 111
	item := model.Item{
		SkuId: 123,
		Name:  "test",
		Count: 1,
		Price: 100,
	}
	repo := NewCartRepository()
	err := repo.AddItemToCart(userId, &item)
	require.NoError(t, err)

	// act
	err = repo.DeleteAllFromCart(userId)
	require.NoError(t, err)
	actual, err := repo.GetItems(userId)

	// assert
	assert.Equal(t, len(actual.Items), 0)
	assert.Equal(t, actual.TotalPrice, uint32(0))
}

func TestDeleteFromCart(t *testing.T) {
	// arrange
	var userId int64 = 111
	item1 := model.Item{
		SkuId: 1,
		Name:  "item 1",
		Count: 1,
		Price: 100,
	}
	item2 := model.Item{
		SkuId: 2,
		Name:  "item 2",
		Count: 1,
		Price: 200,
	}
	repo := NewCartRepository()
	err := repo.AddItemToCart(userId, &item1)
	require.NoError(t, err)
	err = repo.AddItemToCart(userId, &item2)
	require.NoError(t, err)

	// act
	err = repo.DeleteFromCart(userId, item1.SkuId)
	require.NoError(t, err)
	actual, err := repo.GetItems(userId)

	// assert
	assert.Equal(t, actual.Items[item2.SkuId], &item2)
	assert.Equal(t, actual.TotalPrice, item2.Price)
}

func TestGetItems_ReturnEmpty(t *testing.T) {
	// arrange
	var userId int64 = 111

	// act
	repo := NewCartRepository()
	actual, err := repo.GetItems(userId)
	require.NoError(t, err)

	// assert
	assert.Equal(t, len(actual.Items), 0)
	assert.Equal(t, actual.TotalPrice, uint32(0))
}
