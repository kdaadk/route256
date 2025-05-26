package model

type Order struct {
	OrderId int64
	UserId  int64
	Items   []Item
	Status  OrderStatus
}

type Item struct {
	SkuId int64
	Count uint32
}

type Stock struct {
	SkuId      int64
	TotalCount uint32
	Reserved   uint32
}

var (
	OrderStatus_New             OrderStatus = "new"
	OrderStatus_AwaitingPayment OrderStatus = "awaiting_payment"
	OrderStatus_Failed          OrderStatus = "failed"
	OrderStatus_Paid            OrderStatus = "paid"
	OrderStatus_Cancelled       OrderStatus = "cancelled"
)

type OrderStatus string
