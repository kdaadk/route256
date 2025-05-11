package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gojuno/minimock/v3"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"route256/cart/internal/model"
	"route256/cart/internal/service"
	"route256/cart/internal/service/mocks"
)

func setupMocks(mc *minimock.Controller) (*mocks.CartRepoMock, *mocks.ProductClientMock, *Handler) {
	repo := mocks.NewCartRepoMock(mc)
	prod := mocks.NewProductClientMock(mc)
	s := service.NewService(repo, prod)
	h := NewHandler(s)
	return repo, prod, h
}

func newProduct(skuId int64) *model.Product {
	return &model.Product{
		Id:    skuId,
		Name:  fmt.Sprintf("Product %d", skuId),
		Price: 100,
	}
}

func newItem(product *model.Product, count uint16) *model.Item {
	return &model.Item{
		SkuId: product.Id,
		Name:  product.Name,
		Price: product.Price * uint32(count),
		Count: count,
	}
}

func newUserData(items []*model.Item) *model.UserData {
	totalPrice := uint32(0)
	itemsMap := make(map[int64]*model.Item, len(items))
	for _, item := range items {
		gotItem, ok := itemsMap[item.SkuId]
		if !ok {
			itemsMap[item.SkuId] = item
		} else {
			gotItem.Price += item.Price
			gotItem.Count += item.Count
			itemsMap[item.SkuId] = gotItem
		}

		totalPrice += item.Price
	}
	return &model.UserData{
		Items:      itemsMap,
		TotalPrice: totalPrice,
	}
}

func newAddItemRequest(userId, skuId int64, count uint16) *http.Request {
	url := fmt.Sprintf("/user/%d/cart/%d", userId, skuId)
	body, _ := json.Marshal(model.AddToCartRequest{Count: count})
	return httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
}

func newGetCartRequest(userId int64) *http.Request {
	url := fmt.Sprintf("/user/%d/cart", userId)
	return httptest.NewRequest(http.MethodGet, url, nil)
}

func newDeleteAllRequest(userId int64) *http.Request {
	url := fmt.Sprintf("/user/%d/cart", userId)
	return httptest.NewRequest(http.MethodDelete, url, nil)
}

func newDeleteFromCartRequest(userId, skuId int64) *http.Request {
	url := fmt.Sprintf("/user/%d/cart/%d", userId, skuId)
	return httptest.NewRequest(http.MethodDelete, url, nil)
}

func assertStatus(t *testing.T, w *httptest.ResponseRecorder, expected int) {
	t.Helper()
	if actual := w.Result().StatusCode; actual != expected {
		t.Errorf("unexpected status code: got %d, want %d", actual, expected)
	}
}

func assertCart(t *testing.T, router *mux.Router, userId int64, expected *model.UserData) {
	t.Helper()
	w := httptest.NewRecorder()
	router.ServeHTTP(w, newGetCartRequest(userId))

	resp := w.Result()
	defer resp.Body.Close()
	assertStatus(t, w, http.StatusOK)

	var got model.GetItemsFromCartResponse
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	assert.Equal(t, got.TotalPrice, expected.TotalPrice)
	assert.Equal(t, len(got.Items), len(expected.Items))
	for _, gotItem := range got.Items {
		assert.Equal(t, expected.Items[gotItem.SkuId], &gotItem)
	}
}

func newTestRouter(h *Handler) *mux.Router {
	r := mux.NewRouter()
	h.RegisterRoutes(r)
	return r
}

// --- Tests ---
func TestAddToCartHandler_Success(t *testing.T) {
	// arrange
	const (
		userId int64  = 31337
		skuId  int64  = 1076963
		count  uint16 = 1
	)
	product := newProduct(skuId)
	item := newItem(product, count)
	userData := newUserData([]*model.Item{item})

	ctrl := minimock.NewController(t)
	cartRepoMock, productCliMock, h := setupMocks(ctrl)
	productCliMock.GetProductMock.Expect(skuId).Return(product, nil)
	cartRepoMock.AddItemToCartMock.Expect(userId, item).Return(nil)
	cartRepoMock.GetItemsMock.Expect(userId).Return(userData, nil)

	// act
	router := newTestRouter(h)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, newAddItemRequest(userId, skuId, count))

	// assert
	assertStatus(t, w, http.StatusOK)
	assertCart(t, router, userId, userData)
}

func TestAddToCartHandler_SeveralSameAddProduct_ShouldSum(t *testing.T) {
	// arrange
	const (
		userId int64  = 31337
		skuId  int64  = 1076963
		count1 uint16 = 1
		count2 uint16 = 5
	)
	product := newProduct(skuId)
	item1 := newItem(product, count1)
	item2 := newItem(product, count2)
	userData := newUserData([]*model.Item{item1, item2})

	ctrl := minimock.NewController(t)
	cartRepoMock, productCliMock, h := setupMocks(ctrl)
	productCliMock.GetProductMock.Expect(skuId).Return(product, nil)
	cartRepoMock.AddItemToCartMock.Set(func(uId int64, item *model.Item) error {
		if uId != userId {
			t.Errorf("unexpected userId: got %d, want %d", uId, userId)
		}
		if item.SkuId != skuId {
			t.Errorf("unexpected skuId: got %d, want %d", item.SkuId, skuId)
		}
		if item.Count != count1 && item.Count != count2 {
			t.Errorf("unexpected count: got %d, want %d or %d", item.Count, count1, count2)
		}
		return nil
	})
	cartRepoMock.GetItemsMock.Expect(userId).Return(userData, nil)

	// act
	router := newTestRouter(h)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, newAddItemRequest(userId, skuId, count1))
	assertStatus(t, w, http.StatusOK)

	w = httptest.NewRecorder()
	router.ServeHTTP(w, newAddItemRequest(userId, skuId, count2))
	assertStatus(t, w, http.StatusOK)

	// assert
	assertCart(t, router, userId, userData)
}

func TestAddToCartHandler_SeveralDiffAddProduct_ShouldStored(t *testing.T) {
	// arrange
	const (
		userId int64  = 31337
		skuId1 int64  = 1076963
		skuId2 int64  = 1076964
		count  uint16 = 1
	)
	product1 := newProduct(skuId1)
	product2 := newProduct(skuId2)
	item1 := newItem(product1, count)
	item2 := newItem(product2, count)
	userData := newUserData([]*model.Item{item1, item2})

	ctrl := minimock.NewController(t)
	cartRepoMock, productCliMock, h := setupMocks(ctrl)
	productCliMock.GetProductMock.Set(func(sku int64) (*model.Product, error) {
		switch sku {
		case skuId1:
			return product1, nil
		case skuId2:
			return product2, nil
		default:
			t.Errorf("unexpected SKU ID: %d", sku)
			return nil, fmt.Errorf("unknown product")
		}
	})
	cartRepoMock.AddItemToCartMock.Set(func(uId int64, item *model.Item) error {
		if uId != userId {
			t.Errorf("unexpected userId: got %d, want %d", uId, userId)
		}
		if item.SkuId != skuId1 && item.SkuId != skuId2 {
			t.Errorf("unexpected skuId: got %d, want %d or %d", item.Count, skuId1, skuId2)
		}
		if item.Count != count {
			t.Errorf("unexpected count: got %d, want %d", item.Count, count)
		}
		return nil
	})
	cartRepoMock.GetItemsMock.Expect(userId).Return(userData, nil)

	// act
	router := newTestRouter(h)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, newAddItemRequest(userId, skuId1, count))
	assertStatus(t, w, http.StatusOK)

	w = httptest.NewRecorder()
	router.ServeHTTP(w, newAddItemRequest(userId, skuId2, count))
	assertStatus(t, w, http.StatusOK)

	// assert
	assertCart(t, router, userId, userData)
}

func TestAddToCartHandler_InvalidInput_Failure(t *testing.T) {
	cases := []struct {
		name   string
		userId int64
		skuId  int64
		count  uint16
	}{
		{"Zero user ID", 0, 1076963, 1},
		{"Negative user ID", -42, 1076963, 1},
		{"Zero SKU ID", 31337, 0, 1},
		{"Negative SKU ID", 31337, -123, 1},
		{"Zero count", 31337, 1076963, 0},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := minimock.NewController(t)
			_, _, h := setupMocks(ctrl)
			router := newTestRouter(h)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, newAddItemRequest(tc.userId, tc.skuId, tc.count))

			assertStatus(t, w, http.StatusBadRequest)
		})
	}
}

func TestAddToCartHandler_UnknownProduct_Failure(t *testing.T) {
	// arrange
	const (
		userId int64  = 31337
		skuId  int64  = 666666
		count  uint16 = 1
	)

	ctrl := minimock.NewController(t)
	_, productCliMock, h := setupMocks(ctrl)
	productCliMock.GetProductMock.Expect(skuId).Return(nil, model.ErrPreconditionFailed)

	// act
	router := newTestRouter(h)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, newAddItemRequest(userId, skuId, count))

	// assert
	assertStatus(t, w, http.StatusPreconditionFailed)
}

func TestDeleteAllHandler_Success(t *testing.T) {
	// arrange
	const (
		userId int64  = 31337
		skuId  int64  = 1076963
		count  uint16 = 1
	)
	product := newProduct(skuId)
	item := newItem(product, count)
	userData := newUserData([]*model.Item{})

	ctrl := minimock.NewController(t)
	cartRepoMock, productCliMock, h := setupMocks(ctrl)
	productCliMock.GetProductMock.Expect(skuId).Return(product, nil)
	cartRepoMock.AddItemToCartMock.Expect(userId, item).Return(nil)
	cartRepoMock.DeleteAllFromCartMock.Expect(userId).Return(nil)
	cartRepoMock.GetItemsMock.Expect(userId).Return(userData, nil)

	router := newTestRouter(h)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, newAddItemRequest(userId, skuId, count))

	// act
	w = httptest.NewRecorder()
	router.ServeHTTP(w, newDeleteAllRequest(userId))

	// assert
	assertStatus(t, w, http.StatusOK)
	assertCart(t, router, userId, userData)
}

func TestDeleteItemHandler_Success(t *testing.T) {
	// arrange
	const (
		userId int64  = 31337
		skuId  int64  = 1076963
		count  uint16 = 1
	)
	product := newProduct(skuId)
	item := newItem(product, count)
	userData := newUserData([]*model.Item{})

	ctrl := minimock.NewController(t)
	cartRepoMock, productCliMock, h := setupMocks(ctrl)
	productCliMock.GetProductMock.Expect(skuId).Return(product, nil)
	cartRepoMock.AddItemToCartMock.Expect(userId, item).Return(nil)
	cartRepoMock.DeleteFromCartMock.Expect(userId, skuId).Return(nil)
	cartRepoMock.GetItemsMock.Expect(userId).Return(userData, nil)

	router := newTestRouter(h)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, newAddItemRequest(userId, skuId, count))

	// act
	w = httptest.NewRecorder()
	router.ServeHTTP(w, newDeleteFromCartRequest(userId, skuId))

	// assert
	assertStatus(t, w, http.StatusOK)
	assertCart(t, router, userId, userData)
}

func TestGetCartHandler_InvalidUser_Failure(t *testing.T) {
	// arrange
	const userId int64 = 0
	ctrl := minimock.NewController(t)
	_, _, h := setupMocks(ctrl)

	// act
	router := newTestRouter(h)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, newGetCartRequest(userId))

	// assert
	assertStatus(t, w, http.StatusBadRequest)
}
