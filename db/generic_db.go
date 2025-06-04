package db

import (
	"ermes-central-repo/config"
	"github.com/redis/go-redis/v9"
)

const GROUP_LIST_KEY = "existing_groups"

type GenericDB interface {
	AddGroupToUser(groupName string, userId string) error
	RemoveGroupFromUser(groupName string, userId string) error
	GetUserGroups(userId string) ([]string, error)
	CreateGroup(groupName string, participantIDs []string) error
	DeleteGroup(groupName string) error
	GetExistingGroups() ([]string, error)
	Close() error
}

func NewRedisClient() GenericDB {
	rdb := redis.NewClient(&redis.Options{
		Addr:     config.RedisAddr() + ":" + config.RedisPort(),
		Password: config.RedisPassword(),
		DB:       config.RedisDB(),
	})

	return &RedisClient{client: rdb}
}
