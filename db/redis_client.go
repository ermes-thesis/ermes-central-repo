package db

import (
	"ermes-central-repo/config"
	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	client *redis.Client
}

func NewRedisClient() *RedisClient {
	rdb := redis.NewClient(&redis.Options{
		Addr:     config.RedisAddr() + ":" + config.RedisPort(),
		Password: config.RedisPassword(),
		DB:       config.RedisDB(),
	})

	return &RedisClient{client: rdb}
}
