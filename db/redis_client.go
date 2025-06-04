package db

import (
	"context"
	"errors"
	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	client *redis.Client
}

func (r *RedisClient) AddGroupToUser(groupName string, userId string) error {
	ctx := context.Background()

	result, err := r.client.SIsMember(ctx, GROUP_LIST_KEY, groupName).Result()
	if err != nil {
		return err
	}
	if !result {
		return errors.New("group does not exist")
	}

	_, err = r.client.SAdd(ctx, userId, groupName).Result()
	if err != nil {
		return err
	}
	return nil
}

func (r *RedisClient) RemoveGroupFromUser(groupName string, userId string) error {
	ctx := context.Background()
	_, err := r.client.SRem(ctx, userId, groupName).Result()
	if err != nil {
		return err
	}
	return nil
}

func (r *RedisClient) GetUserGroups(userId string) ([]string, error) {
	ctx := context.Background()
	return r.client.SMembers(ctx, userId).Result()
}

func (r *RedisClient) CreateGroup(groupName string, participantIDs []string) error {
	ctx := context.Background()
	_, err := r.client.SAdd(ctx, GROUP_LIST_KEY, groupName).Result()
	if err != nil {
		return err
	}

	for _, participantID := range participantIDs {
		_, err = r.client.SAdd(ctx, participantID, groupName).Result()
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *RedisClient) DeleteGroup(groupName string) error {
	ctx := context.Background()

	var cursor uint64
	for {
		var keys []string
		var err error
		keys, cursor, err = r.client.Scan(ctx, cursor, "*", 10).Result()
		if err != nil {
			return err
		}

		for _, key := range keys {
			_, err = r.client.SRem(ctx, key, groupName).Result()
			if err != nil {
				return err
			}
		}

		if cursor == 0 {
			break
		}
	}

	return nil
}

func (r *RedisClient) GetExistingGroups() ([]string, error) {
	ctx := context.Background()
	return r.client.SMembers(ctx, GROUP_LIST_KEY).Result()
}

func (r *RedisClient) Close() error {
	return r.client.Close()
}
