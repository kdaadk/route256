package model

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type Order struct {
	OrderId   int64            `db:"id" json:"orderId"`
	UserId    int64            `db:"user_id" json:"userId"`
	Items     []Item           `db:"-" json:"items"`
	Status    OrderStatus      `db:"status" json:"status"`
	CreatedAt pgtype.Timestamp `db:"created_at" json:"createdAt"`
	UpdatedAt pgtype.Timestamp `db:"updated_at" json:"updatedAt"`
}

type Item struct {
	OrderId int64  `db:"order_id" json:"orderId"`
	SkuId   int64  `db:"sku_id" json:"skuId"`
	Count   uint32 `db:"count" json:"count"`
}

type Stock struct {
	SkuId      int64  `db:"sku_id" json:"skuId"`
	TotalCount uint32 `db:"total_count" json:"totalCount"`
	Reserved   uint32 `db:"reserved" json:"reserved"`
}

const (
	OrderStatus_Unknown OrderStatus = iota
	OrderStatus_New
	OrderStatus_AwaitingPayment
	OrderStatus_Failed
	OrderStatus_Paid
	OrderStatus_Cancelled
)

type OrderStatus int64

type OrderChangedStatusEvent struct {
	OrderId   int64       `db:"id" json:"orderId"`
	NewStatus OrderStatus `db:"status" json:"newStatus"`
}
