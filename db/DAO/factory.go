package DAO

import "github.com/jackc/pgx/v5/pgxpool"

type DAOFactory struct {
	db *pgxpool.Pool
}

func NewDAOFactory(db *pgxpool.Pool) (*DAOFactory, error) {
	return &DAOFactory{db: db}, nil
}

func (f *DAOFactory) User() UserDAO {
	return NewUserDAO(f.db)
}

func (f *DAOFactory) Group() GroupDAO {
	return NewGroupDAO(f.db)
}

func (f *DAOFactory) UserGroupRelation() UserGroupRelDAO {
	return NewUserGroupRelDAO(f.db)
}
