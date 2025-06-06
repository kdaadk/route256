package service

import (
	"context"
	"fmt"
	"github.com/gojuno/minimock/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"route256/cart/internal/model"
	"route256/cart/internal/service/mocks"
	proto "route256/cart/vendor-proto/route256/loms"
	"testing"
)

func setupMocks(mc *minimock.Controller) (*mocks.CartRepoMock, *mocks.ProductClientMock, *mocks.LomsClientMock, *Service) {
	repoMock := mocks.NewCartRepoMock(mc)
	productClientMock := mocks.NewProductClientMock(mc)
	lomsCliMock := mocks.NewLomsClientMock(mc)
	s := NewService(repoMock, productClientMock, lomsCliMock)

	return repoMock, productClientMock, lomsCliMock, s
}

func TestGetProduct(t *testing.T) {
	// arrange
	var productId int64 = 123
	ctx := context.Background()
	wantProduct := model.Product{
		Id:    productId,
		Name:  "test",
		Price: 123,
	}

	ctrl := minimock.NewController(t)
	_, productCliMock, _, serv := setupMocks(ctrl)
	productCliMock.GetProductMock.Expect(ctx, productId).Return(&wantProduct, nil)

	// act
	actual, err := serv.productCli.GetProduct(ctx, productId)
	require.NoError(t, err)

	// assert
	assert.Equal(t, &wantProduct, actual)
}

func TestAddItemToCart(t *testing.T) {
	// arrange
	var userId int64 = 333
	var productId int64 = 123
	product := &model.Product{
		Id:    productId,
		Name:  fmt.Sprintf("Product %d", productId),
		Price: 100,
	}
	ctx := context.Background()
	wantItem := model.DtoItem{
		SkuId: productId,
		Count: 1,
	}

	ctrl := minimock.NewController(t)
	cartRepoMock, productCliMock, lomsCliMock, serv := setupMocks(ctrl)
	productCliMock.GetProductMock.Expect(ctx, productId).Return(product, nil)
	lomsCliMock.GetStockInfosMock.Expect(ctx, []int64{productId}).Return(
		&proto.GetStockInfoResponse{StockInfos: []*proto.StockInfo{{SkuId: productId, Count: uint32(wantItem.Count)}}},
		nil)
	cartRepoMock.AddItemToCartMock.Expect(ctx, userId, &wantItem).Return(nil)

	// act assert
	err := serv.AddItemToCart(ctx, userId, productId, 1)
	require.NoError(t, err)
}

func TestDeleteFromCart(t *testing.T) {
	// arrange
	var userId int64 = 333
	var skuId int64 = 111
	ctx := context.Background()

	ctrl := minimock.NewController(t)
	cartRepoMock, _, _, serv := setupMocks(ctrl)
	cartRepoMock.DeleteFromCartMock.Expect(ctx, userId, skuId).Return(nil)

	// act assert
	err := serv.DeleteFromCart(ctx, userId, skuId)
	require.NoError(t, err)
}

func TestGetItems(t *testing.T) {
	// arrange
	var userId int64 = 333
	ctx := context.Background()
	wantUserData := model.DtoUserData{
		Items: map[int64]*model.DtoItem{},
	}

	ctrl := minimock.NewController(t)
	cartRepoMock, _, _, serv := setupMocks(ctrl)
	cartRepoMock.GetItemsMock.Expect(ctx, userId).Return(&wantUserData, nil)

	// act
	actual, err := serv.GetItems(ctx, userId)
	require.NoError(t, err)

	// assert
	assert.Equal(t, len(wantUserData.Items), len(actual.Items))
	for i, item := range wantUserData.Items {
		actualItem := actual.Items[i]
		assert.Equal(t, item.SkuId, actualItem.SkuId)
		assert.Equal(t, item.Count, actualItem.Count)
	}
}
