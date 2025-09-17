package db

import (
	"ermes-central-repo/config"
	"ermes-central-repo/logger"
	"github.com/redis/go-redis/v9"
)

const GroupListKey = "existing_groups"

type GenericDB interface {
	AddGroupToUser(groupName string, userId string) error
	RemoveGroupFromUser(groupName string, userId string) error
	GetUserGroups(userId string) ([]string, error)
	CreateGroup(groupName string, participantIDs []string) error
	DeleteGroup(groupName string) error
	GetExistingGroups() ([]string, error)
	Close() error
}

func NewRedisClient(cfg *config.Config, log *logger.Logger) GenericDB {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr + ":" + cfg.RedisPort,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	return &RedisClient{client: rdb, log: log}
}
