package model

import "errors"

var (
	ErrPreconditionFailed = errors.New("precondition failed")
)

type UserData struct {
	Items      map[int64]*Item
	TotalPrice uint32
}

type DtoUserData struct {
	Items map[int64]*DtoItem
}

type AddToCartRequest struct {
	Count uint16 `json:"count" validate:"required,min=1"`
}

type GetItemsFromCartResponse struct {
	Items      []Item `json:"items"`
	TotalPrice uint32 `json:"total_price"`
}

type Item struct {
	SkuId int64  `json:"sku_id"`
	Name  string `json:"name"`
	Count uint16 `json:"count"`
	Price uint32 `json:"price"`
}

type DtoItem struct {
	SkuId int64  `json:"sku_id"`
	Count uint16 `json:"count"`
}

type Product struct {
	Id    int64  `json:"id"`
	Name  string `json:"name"`
	Price uint32 `json:"price"`
}

type CreateOrderRequest struct {
	UserId int64       `json:"user_id"`
	Items  []OrderItem `json:"items"`
}

type OrderItem struct {
	SkuId int64  `json:"sku_id"`
	Count uint16 `json:"count"`
}
