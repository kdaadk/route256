package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v5"
	"route256/loms/internal/model"
)

type OrdersRepository struct {
	db *model.DBPools
}

func NewOrdersRepository(db *model.DBPools) *OrdersRepository {
	return &OrdersRepository{
		db: db,
	}
}

func (r *OrdersRepository) CreateOrder(tx pgx.Tx, userId int64, items []*model.Item) (int64, error) {
	ctx := context.Background()
	var orderId int64
	err := tx.QueryRow(ctx,
		`INSERT INTO orders(user_id, status, created_at, updated_at)
     VALUES($1, $2, now(), now())
     RETURNING id`,
		userId,
		model.OrderStatus_New,
	).Scan(&orderId)
	if err != nil {
		return 0, fmt.Errorf("failed to insert order: %w", err)
	}

	for _, item := range items {
		_, err = tx.Exec(ctx,
			`INSERT INTO order_items(order_id, sku_id, count)
             VALUES($1, $2, $3)`,
			orderId,
			item.SkuId,
			item.Count,
		)
		if err != nil {
			return 0, fmt.Errorf("failed to insert order item: %w", err)
		}
	}

	return orderId, nil
}

func (r *OrdersRepository) SetState(tx pgx.Tx, orderId int64, status model.OrderStatus) error {
	ctx := context.Background()
	_, err := tx.Exec(ctx,
		`UPDATE orders
         SET status = $1, updated_at = now()
         WHERE id = $2;`,
		status,
		orderId,
	)
	if err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	return nil
}

func (r *OrdersRepository) GetById(orderId int64) (*model.Order, error) {
	ctx := context.Background()

	const query = `
        SELECT
            o.id,
            o.user_id,
            o.status,
            o.created_at,
            o.updated_at,
            (
                SELECT json_agg(
                    json_build_object(
                        'skuId', oi.sku_id,
                        'count', oi.count
                    )
                )
                FROM order_items oi
                WHERE oi.order_id = o.id
            ) AS items
        FROM orders o
        WHERE o.id = $1
    `

	var order model.Order
	var itemsJSON []byte

	err := r.db.Replica.QueryRow(ctx, query, orderId).Scan(
		&order.OrderId,
		&order.UserId,
		&order.Status,
		&order.CreatedAt,
		&order.UpdatedAt,
		&itemsJSON,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("order not found")
		}
		return nil, fmt.Errorf("failed to get order by id: %w", err)
	}

	// Handle case when no items exist (NULL from database)
	if itemsJSON != nil {
		if err := json.Unmarshal(itemsJSON, &order.Items); err != nil {
			return nil, fmt.Errorf("failed to unmarshal order items: %w", err)
		}
	}

	return &order, nil
}
