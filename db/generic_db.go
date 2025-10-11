package db

import (
	"context"
	"ermes-central-repo/config"

	"github.com/ermes-thesis/ermes-common/logger"
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

func NewRedisClient(cfg *config.Config, log *logger.Logger) (GenericDB, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr + ":" + cfg.RedisPort,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	pong, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		log.Error("Failed to connect to Redis", "error", err)
		return nil, err
	}
	log.Info("Connected to Redis", "response", pong)

	return &RedisClient{client: rdb, log: log}, nil
}
