package db

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	host     = "postgres"
	port     = "5432"
	user     = "user"
	password = "password"
	dbname   = "central_repo"
)

func NewPool() (*pgxpool.Pool, error) {
	ctx := context.Background()
	connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, password, host, port, dbname)
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, err
	}

	config.MaxConns = 10
	config.MinConns = 2

	return pgxpool.NewWithConfig(ctx, config)
}
