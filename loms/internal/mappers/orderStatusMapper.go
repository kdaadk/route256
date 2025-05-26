package mappers

import (
	"route256/loms/internal/model"
	"route256/loms/proto"
)

func MapOrderStatusToBll(status model.OrderStatus) proto.OrderStatus {
	switch status {
	case model.OrderStatus_New:
		return proto.OrderStatus_NEW
	case model.OrderStatus_AwaitingPayment:
		return proto.OrderStatus_AWAITING_PAYMENT
	case model.OrderStatus_Failed:
		return proto.OrderStatus_FAILED
	case model.OrderStatus_Paid:
		return proto.OrderStatus_PAID
	case model.OrderStatus_Cancelled:
		return proto.OrderStatus_CANCELLED
	default:
		return proto.OrderStatus_UNKNOWN
	}
}
