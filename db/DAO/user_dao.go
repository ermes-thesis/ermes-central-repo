package DAO

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserDAO interface {
	Add(user User) error
	GetById(userId int) (*User, error)
	GetByName(userName string) ([]User, error)
	GetBySurname(userSurname string) ([]User, error)
	GetAll() ([]User, error)
}

type userDAOImpl struct {
	pool *pgxpool.Pool
}

func NewUserDAO(pool *pgxpool.Pool) UserDAO {
	return &userDAOImpl{pool: pool}
}

func (u userDAOImpl) Add(user User) error {
	ctx := context.Background()
	query := "INSERT INTO users(id,name,surname) VALUES(@userId, @userName, @userSurname)"
	args := pgx.NamedArgs{
		"userId":      user.Id,
		"userName":    user.Name,
		"userSurname": user.Surname,
	}

	_, err := u.pool.Exec(ctx, query, args)
	if err != nil {
		return err
	}
	return nil
}

func (u userDAOImpl) GetById(userId int) (*User, error) {
	ctx := context.Background()
	query := "SELECT * FROM users WHERE id = @userId"
	args := pgx.NamedArgs{
		"userId": userId,
	}

	row := u.pool.QueryRow(ctx, query, args)

	var user User
	err := row.Scan(&user.Id, &user.Name, &user.Surname)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (u userDAOImpl) GetByName(name string) ([]User, error) {
	ctx := context.Background()
	query := "SELECT * FROM users WHERE name = @userName"
	args := pgx.NamedArgs{
		"userName": name,
	}

	rows, err := u.pool.Query(ctx, query, args)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return pgx.CollectRows(rows, pgx.RowToStructByName[User])
}

func (u userDAOImpl) GetBySurname(surname string) ([]User, error) {
	ctx := context.Background()
	query := "SELECT * FROM users WHERE surname = @surname"
	args := pgx.NamedArgs{
		"surname": surname,
	}

	rows, err := u.pool.Query(ctx, query, args)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return pgx.CollectRows(rows, pgx.RowToStructByName[User])
}

func (u userDAOImpl) GetAll() ([]User, error) {
	ctx := context.Background()
	query := "SELECT * FROM users"

	rows, err := u.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	return pgx.CollectRows(rows, pgx.RowToStructByName[User])
}
