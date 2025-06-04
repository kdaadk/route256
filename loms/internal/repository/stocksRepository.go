package repository

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"route256/loms/internal/model"
)

type StocksRepository struct {
	db *model.DBPools
}

func NewStocksRepository(db *model.DBPools) *StocksRepository {
	return &StocksRepository{db: db}
}

func (r *StocksRepository) Reserve(tx pgx.Tx, items []*model.Item) error {
	ctx := context.Background()
	for _, item := range items {
		res, err := tx.Exec(ctx,
			`UPDATE stocks
         SET reserved = reserved + $1
         WHERE sku = $2
         AND (total_count - reserved) >= $1`,
			item.Count,
			item.SkuId,
		)

		if err != nil {
			return fmt.Errorf("failed to reserve stock: %w", err)
		}

		rowsAffected := res.RowsAffected()
		if rowsAffected == 0 {
			return fmt.Errorf("failed to check sku existence: %w", err)
		}
	}

	return nil
}

func (r *StocksRepository) CancelReserve(tx pgx.Tx, skuId int64, count uint32) error {
	ctx := context.Background()
	res, err := tx.Exec(ctx,
		`UPDATE stocks
         SET reserved = reserved - $1
         WHERE sku = $2
         AND reserved >= $1`,
		count,
		skuId,
	)
	if err != nil {
		return fmt.Errorf("failed to cancel reservation: %w", err)
	}

	rowsAffected := res.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("failed to verify SKU existence: %w", err)
	}

	return nil
}

func (r *StocksRepository) RemoveReserve(tx pgx.Tx, skuId int64, count uint32) error {
	res, err := tx.Exec(context.Background(),
		`UPDATE stocks
         SET
             reserved = reserved - $1,
             total_count = total_count - $1
         WHERE
             sku = $2
             AND reserved >= $1
             AND total_count >= $1`,
		count,
		skuId,
	)

	if err != nil {
		return fmt.Errorf("failed to remove reservation: %w", err)
	}

	rowsAffected := res.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("failed to check stock levels: %w", err)
	}

	return nil
}

func (r *StocksRepository) GetStockInfos(skuIds []int64) ([]*model.Stock, error) {
	if len(skuIds) == 0 {
		return nil, nil
	}

	ctx := context.Background()
	query := `SELECT sku, total_count, reserved FROM stocks WHERE sku = ANY($1)`
	rows, err := r.db.Replica.Query(ctx, query, skuIds)
	if err != nil {
		return nil, fmt.Errorf("failed to query stocks: %w", err)
	}
	defer rows.Close()

	var stocks []*model.Stock
	for rows.Next() {
		var stock model.Stock
		if err := rows.Scan(&stock.SkuId, &stock.TotalCount, &stock.Reserved); err != nil {
			return nil, fmt.Errorf("failed to scan stock row: %w", err)
		}
		stocks = append(stocks, &stock)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return stocks, nil
}
