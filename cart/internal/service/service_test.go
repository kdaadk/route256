package service

import (
	"github.com/gojuno/minimock/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"route256/cart/internal/model"
	"route256/cart/internal/service/mocks"
	"testing"
)

func setupMocks(mc *minimock.Controller) (*mocks.CartRepoMock, *mocks.ProductClientMock, *Service) {
	repoMock := mocks.NewCartRepoMock(mc)
	productClientMock := mocks.NewProductClientMock(mc)
	s := NewService(repoMock, productClientMock)

	return repoMock, productClientMock, s
}

func TestGetProduct(t *testing.T) {
	// arrange
	var productId int64 = 123
	wantProduct := model.Product{
		Id:    productId,
		Name:  "test",
		Price: 123,
	}

	ctrl := minimock.NewController(t)
	_, productCliMock, serv := setupMocks(ctrl)
	productCliMock.GetProductMock.Expect(productId).Return(&wantProduct, nil)

	// act
	actual, err := serv.GetProduct(productId)
	require.NoError(t, err)

	// assert
	assert.Equal(t, &wantProduct, actual)
}

func TestAddItemToCart(t *testing.T) {
	// arrange
	var userId int64 = 333
	var productId int64 = 123
	wantItem := model.Item{
		SkuId: productId,
		Name:  "test",
		Price: 123,
		Count: 1,
	}

	ctrl := minimock.NewController(t)
	cartRepoMock, _, serv := setupMocks(ctrl)
	cartRepoMock.AddItemToCartMock.Expect(userId, &wantItem).Return(nil)

	// act assert
	err := serv.AddItemToCart(userId, &wantItem)
	require.NoError(t, err)
}

func TestDeleteAllFromCart(t *testing.T) {
	// arrange
	var userId int64 = 333

	ctrl := minimock.NewController(t)
	cartRepoMock, _, serv := setupMocks(ctrl)
	cartRepoMock.DeleteAllFromCartMock.Expect(userId).Return(nil)

	// act assert
	err := serv.DeleteAllFromCart(userId)
	require.NoError(t, err)
}

func TestDeleteFromCart(t *testing.T) {
	// arrange
	var userId int64 = 333
	var skuId int64 = 111

	ctrl := minimock.NewController(t)
	cartRepoMock, _, serv := setupMocks(ctrl)
	cartRepoMock.DeleteFromCartMock.Expect(userId, skuId).Return(nil)

	// act assert
	err := serv.DeleteFromCart(userId, skuId)
	require.NoError(t, err)
}

func TestGetItems(t *testing.T) {
	// arrange
	var userId int64 = 333
	wantUserData := model.UserData{
		Items:      map[int64]*model.Item{},
		TotalPrice: 0,
	}

	ctrl := minimock.NewController(t)
	cartRepoMock, _, serv := setupMocks(ctrl)
	cartRepoMock.GetItemsMock.Expect(userId).Return(&wantUserData, nil)

	// act
	actual, err := serv.GetItems(userId)
	require.NoError(t, err)

	// assert
	assert.Equal(t, &wantUserData, actual)
}
