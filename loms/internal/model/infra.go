package model

import "github.com/jackc/pgx/v5/pgxpool"

type DBPools struct {
	Master  *pgxpool.Pool
	Replica *pgxpool.Pool
}
