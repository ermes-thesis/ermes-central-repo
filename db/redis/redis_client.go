package redis

import (
	"centralRepoDemo/db"
	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	client *redis.Client
}

func NewRedisClient() *RedisClient {
	rdbConfigs, _ := db.GetRedisConfig()
	rdb := redis.NewClient(&redis.Options{
		Addr:     rdbConfigs.Addr + ":" + rdbConfigs.Port,
		Password: rdbConfigs.Password,
		DB:       rdbConfigs.DB,
	})

	return &RedisClient{client: rdb}
}
