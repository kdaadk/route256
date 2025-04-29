package model

import "errors"

var (
	ErrPreconditionFailed = errors.New("precondition failed")
)

type UserData struct {
	Items      map[int64]*Item
	TotalPrice uint32
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

type Product struct {
	Id    int64  `json:"id"`
	Name  string `json:"name"`
	Price uint32 `json:"price"`
}
