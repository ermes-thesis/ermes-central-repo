package DAO

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type GroupDAO interface {
	Add(user Group) error
	GetByName(name string) (*Group, error)
	GetAll() ([]Group, error)
}

type groupDAOImpl struct {
	pool *pgxpool.Pool
}

func NewGroupDAO(pool *pgxpool.Pool) GroupDAO {
	return &groupDAOImpl{pool: pool}
}

func (g groupDAOImpl) Add(group Group) error {
	ctx := context.Background()
	query := "INSERT INTO groups(group_name) VALUES(@gName)"
	args := pgx.NamedArgs{
		"gName": group.GroupName,
	}

	_, err := g.pool.Exec(ctx, query, args)
	if err != nil {
		return err
	}
	return nil
}

func (g groupDAOImpl) GetByName(name string) (*Group, error) {
	ctx := context.Background()
	query := "SELECT * FROM groups WHERE group_name = @gName"
	args := pgx.NamedArgs{
		"gName": name,
	}

	rows, err := g.pool.Query(ctx, query, args)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	group, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Group])
	return &group, err
}

func (g groupDAOImpl) GetAll() ([]Group, error) {
	ctx := context.Background()
	query := "SELECT * FROM groups"

	rows, err := g.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	return pgx.CollectRows(rows, pgx.RowToStructByName[Group])
}
