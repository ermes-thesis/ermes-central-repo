package DAO

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserGroupRelDAO interface {
	Add(user UserGroupRelation) error
	GetByUserId(userId int) ([]UserGroupRelation, error)
	GetByGroupName(groupName string) ([]UserGroupRelation, error)
	GetAll() ([]UserGroupRelation, error)
}

type userGroupRelDAOImpl struct {
	pool *pgxpool.Pool
}

func NewUserGroupRelDAO(pool *pgxpool.Pool) UserGroupRelDAO {
	return &userGroupRelDAOImpl{pool: pool}
}

func (ugr userGroupRelDAOImpl) Add(userGroupRel UserGroupRelation) error {
	ctx := context.Background()
	query := "INSERT INTO users_groups(user_id, assigned_group) VALUES(@uId,@gName)"
	args := pgx.NamedArgs{
		"uId":   userGroupRel.UserId,
		"gName": userGroupRel.AssignedGroup,
	}

	_, err := ugr.pool.Exec(ctx, query, args)
	if err != nil {
		return err
	}
	return nil
}

func (ugr userGroupRelDAOImpl) GetByUserId(userId int) ([]UserGroupRelation, error) {
	ctx := context.Background()
	fmt.Println(userId)
	query := "SELECT * FROM users_groups WHERE user_id = 2"
	//args := pgx.NamedArgs{
	//	"uId": userId,
	//}

	rows, err := ugr.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return pgx.CollectRows(rows, pgx.RowToStructByName[UserGroupRelation])
}

func (ugr userGroupRelDAOImpl) GetByGroupName(groupName string) ([]UserGroupRelation, error) {
	ctx := context.Background()
	query := "SELECT * FROM users_groups WHERE assigned_group = @gName"
	args := pgx.NamedArgs{
		"gName": groupName,
	}

	rows, err := ugr.pool.Query(ctx, query, args)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return pgx.CollectRows(rows, pgx.RowToStructByName[UserGroupRelation])
}

func (ugr userGroupRelDAOImpl) GetAll() ([]UserGroupRelation, error) {
	ctx := context.Background()
	query := "SELECT * FROM users_groups"

	rows, err := ugr.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	return pgx.CollectRows(rows, pgx.RowToStructByName[UserGroupRelation])
}
