package postgres

import (
	"centralRepoDemo/db"
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPostgresPool() (*pgxpool.Pool, error) {
	ctx := context.Background()
	pgConf, _ := db.GetPostgresConfig()

	connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		pgConf.Host, pgConf.Password, pgConf.Host, pgConf.Port, pgConf.DBname)
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, err
	}

	config.MaxConns = 10
	config.MinConns = 2

	return pgxpool.NewWithConfig(ctx, config)
}
