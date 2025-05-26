package repository

import (
	"errors"
	"fmt"
	"route256/loms/internal/model"
)

type StocksRepository struct {
	stocks map[int64]*model.Stock
}

func NewStocksRepository() *StocksRepository {
	return &StocksRepository{
		stocks: map[int64]*model.Stock{
			773297411: {SkuId: 773297411, TotalCount: 150, Reserved: 10},
			1002:      {SkuId: 1002, TotalCount: 200, Reserved: 20},
			1003:      {SkuId: 1003, TotalCount: 250, Reserved: 30},
			1004:      {SkuId: 1004, TotalCount: 300, Reserved: 40},
			1005:      {SkuId: 1005, TotalCount: 350, Reserved: 50},
		},
	}
}

func (r *StocksRepository) Reserve(skuId int64, count uint32) (model.OrderStatus, error) {
	stock, ok := r.stocks[skuId]
	if !ok {
		return model.OrderStatus_Failed, errors.New("sku not found")
	}

	if count > (stock.TotalCount - stock.Reserved) {
		return model.OrderStatus_Failed, errors.New("total count exceeded")
	}

	stock.Reserved += count
	r.stocks[skuId] = stock
	return model.OrderStatus_AwaitingPayment, nil
}

func (r *StocksRepository) CancelReserve(skuId int64, count uint32) error {
	stock, ok := r.stocks[skuId]
	if !ok {
		return errors.New("sku not found")
	}

	stock.Reserved -= count
	r.stocks[skuId] = stock
	return nil
}

func (r *StocksRepository) RemoveReserve(skuId int64, count uint32) error {
	stock, ok := r.stocks[skuId]
	if !ok {
		return errors.New("sku not found")
	}

	stock.Reserved -= count
	stock.TotalCount -= count
	r.stocks[skuId] = stock
	return nil
}

func (r *StocksRepository) GetStockInfos(skuIds []int64) (*[]model.Stock, error) {
	stocks := make([]model.Stock, 0)
	for _, skuId := range skuIds {
		stock, ok := r.stocks[skuId]
		if !ok {
			return nil, errors.New(fmt.Sprintf("sku %d not found", skuId))
		}

		stocks = append(stocks, *stock)
	}

	return &stocks, nil
}
