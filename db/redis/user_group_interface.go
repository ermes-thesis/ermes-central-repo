package redis

import (
	"context"
	"errors"
	"strconv"
)

const GROUP_LIST_KEY = "existing_groups"

func (r *RedisClient) AddGroupToUser(groupName string, userId int) error {
	ctx := context.Background()

	result, err := r.client.SIsMember(ctx, GROUP_LIST_KEY, groupName).Result()
	if err != nil {
		return err
	}
	if !result {
		return errors.New("group does not exist")
	}

	_, err = r.client.SAdd(ctx, strconv.Itoa(userId), groupName).Result()
	if err != nil {
		return err
	}
	return nil
}

func (r *RedisClient) RemoveGroupFromUser(groupName string, userId int) error {
	ctx := context.Background()
	_, err := r.client.SRem(ctx, strconv.Itoa(userId), groupName).Result()
	if err != nil {
		return err
	}
	return nil
}

func (r *RedisClient) GetUserGroups(userId int) ([]string, error) {
	ctx := context.Background()
	return r.client.SMembers(ctx, strconv.Itoa(userId)).Result()
}

func (r *RedisClient) CreateGroup(groupName string) error {
	ctx := context.Background()
	_, err := r.client.SAdd(ctx, GROUP_LIST_KEY, groupName).Result()
	if err != nil {
		return err
	}
	return nil
}

func (r *RedisClient) DeleteGroup(groupName string) error {
	ctx := context.Background()
	_, err := r.client.SRem(ctx, GROUP_LIST_KEY, groupName).Result()
	if err != nil {
		return err
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
