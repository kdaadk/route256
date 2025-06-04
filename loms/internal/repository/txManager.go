package repository

import (
	"context"
	"github.com/jackc/pgx/v5"
	"route256/loms/internal/model"
)

type TxManager struct {
	db *model.DBPools
}

func NewTxManager(db *model.DBPools) *TxManager {
	return &TxManager{db: db}
}

func (m *TxManager) Tx(ctx context.Context, fn func(ctx context.Context, tx pgx.Tx) error, opts *pgx.TxOptions) error {
	if opts == nil {
		opts = new(pgx.TxOptions)
	}

	tx, err := m.db.Master.BeginTx(ctx, *opts)
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		} else if err != nil {
			_ = tx.Rollback(ctx)
		} else {
			err = tx.Commit(ctx)
		}
	}()

	err = fn(ctx, tx)
	return err
}
