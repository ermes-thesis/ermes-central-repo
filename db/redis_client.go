package db

import (
	"context"
	"errors"
	"fmt"
	"github.com/ermes-thesis/ermes-common/logger"
	"github.com/redis/go-redis/v9"
	"strings"
)

type RedisClient struct {
	client *redis.Client
	log    *logger.Logger
}

func (r *RedisClient) AddGroupToUser(groupName string, userId string) error {
	ctx := context.Background()

	userKey := "data:" + userId

	exists, err := r.client.SIsMember(ctx, GroupListKey, groupName).Result()
	if err != nil {
		r.log.Error("Failed to check if group exists",
			"group_name", groupName,
			"user_id", userId,
			"error", err,
		)
		return fmt.Errorf("failed to check if group exists: %v", err)
	}
	if !exists {
		r.log.Warn("Attempted to add user to non-existing group",
			"group_name", groupName,
			"user_id", userId,
		)
		return errors.New("group does not exist")
	}

	if _, err = r.client.SAdd(ctx, userKey, groupName).Result(); err != nil {
		r.log.Error("Failed to add user to group",
			"group_name", groupName,
			"user_id", userId,
			"error", err,
		)
		return fmt.Errorf("failed to add user to group: %v", err)
	}

	r.log.Infof("Successfully added '%s' user to '%s' group", userId, groupName)
	return nil
}

func (r *RedisClient) RemoveGroupFromUser(groupName string, userId string) error {
	ctx := context.Background()
	userKey := "data:" + userId
	_, err := r.client.SRem(ctx, userKey, groupName).Result()
	if err != nil {
		r.log.Error("Failed to remove user from group",
			"group_name", groupName,
			"user_id", userId,
			"error", err,
		)
		return fmt.Errorf("failed to remove user from group: %v", err)
	}

	r.log.Infof("Successfully removed '%s' user from '%s' group", userId, groupName)
	return nil
}

func (r *RedisClient) GetUserGroups(userId string) ([]string, error) {
	ctx := context.Background()
	userKey := "data:" + userId
	groups, err := r.client.SMembers(ctx, userKey).Result()
	if err != nil {
		r.log.Error("Failed to retrieve user groups",
			"user_id", userId,
			"error", err,
		)
		return nil, fmt.Errorf("failed to retrieve user groups: %v", err)
	}

	r.log.Infof("Retrieved user '%s' groups successfully: %v", userId, groups)
	return groups, nil
}

func (r *RedisClient) CreateGroup(groupName string, participantIDs []string) error {
	ctx := context.Background()

	_, err := r.client.SAdd(ctx, GroupListKey, groupName).Result()
	if err != nil {
		r.log.Error("Failed to create group",
			"group_name", groupName,
			"participant_ids", participantIDs,
			"error", err,
		)
		return fmt.Errorf("failed to create group: %v", err)
	}

	for _, participantID := range participantIDs {
		_, err = r.client.SAdd(ctx, "data:"+participantID, groupName).Result()
		if err != nil {
			r.log.Error("Failed to add participant to group",
				"group_name", groupName,
				"participant_id", participantID,
				"error", err,
			)
			return fmt.Errorf("failed to add participant '%s' to group '%s': %v", participantID, groupName, err)
		}
	}

	r.log.Info("Group created successfully",
		"group_name", groupName,
		"participant_ids", participantIDs,
	)
	return nil
}

func (r *RedisClient) DeleteGroup(groupName string) error {
	ctx := context.Background()

	var cursor uint64
	for {
		var userIDs []string
		var err error
		userIDs, cursor, err = r.client.Scan(ctx, cursor, "data:*", 100).Result()
		if err != nil {
			r.log.Error("Failed to scan userIDs during group deletion",
				"group_name", groupName,
				"error", err)
			r.log.Debug("Failed to scan userIDs during group deletion",
				"cursor", cursor,
				"userIDs", userIDs,
			)
			return fmt.Errorf("failed to scan userIDs for deletion: %v", err)
		}

		for _, userID := range userIDs {
			_, err = r.client.SRem(ctx, userID, groupName).Result()
			if err != nil {
				r.log.Error("Failed to remove user from group",
					"group_name", groupName,
					"user_id", userID,
					"error", err,
				)
				return fmt.Errorf("failed to remove user '%s' from group '%s': %v", userID, groupName, err)
			}
		}

		if cursor == 0 {
			break
		}
	}

	_, err := r.client.SRem(ctx, GroupListKey, groupName).Result()
	if err != nil {
		r.log.Error("Failed to remove group from existing ones")
	}

	r.log.Infof("Group '%s' deleted successfully", groupName)
	return nil
}

func (r *RedisClient) GetExistingGroups() ([]string, error) {
	ctx := context.Background()
	groups, err := r.client.SMembers(ctx, GroupListKey).Result()
	if err != nil {
		r.log.Error("Failed to retrieve existing groups",
			"error", err)
		return nil, fmt.Errorf("failed to retrieve existing groups: %v", err)
	}

	r.log.Infof("Retrieved existing groups successfully: %v", groups)
	return groups, nil
}

func (r *RedisClient) GetGroupUsers(groupName string) ([]string, error) {
	ctx := context.Background()

	var usersInGroup []string

	var cursor uint64
	for {
		var userIDs []string
		var err error
		userIDs, cursor, err = r.client.Scan(ctx, cursor, "data:*", 100).Result()
		if err != nil {
			r.log.Error("Failed to scan userIDs during group participant retrieval",
				"group_name", groupName,
				"error", err)
			r.log.Debug("Failed to scan userIDs during group participant retrieval",
				"cursor", cursor,
				"userIDs", userIDs,
			)
			return nil, fmt.Errorf("failed to scan userIDs participant retrieval: %v", err)
		}

		pipe := r.client.Pipeline()
		var cmds []*redis.BoolCmd

		for _, userId := range userIDs {
			cmds = append(cmds, pipe.SIsMember(ctx, userId, groupName))
		}

		_, err = pipe.Exec(ctx)
		if err != nil && !errors.Is(err, redis.Nil) {
			r.log.Error("Failed to execute pipeline during group participant retrieval",
				"group_name", groupName,
				"error", err)
			return nil, fmt.Errorf("failed to check group membership: %v", err)
		}

		for index, cmd := range cmds {
			isIn, err := cmd.Result()
			if err != nil {
				r.log.Error("Failed to check if a user was in a group",
					"group_name", groupName,
					"user_id", userIDs[index],
					"error", err,
				)
				return nil, fmt.Errorf("failed to remove user '%s' from group '%s': %v", userIDs[index], groupName, err)
			}
			if isIn {
				usersInGroup = append(usersInGroup, strings.TrimPrefix(userIDs[index], "data:"))
			}
		}

		if cursor == 0 {
			break
		}
	}
	return usersInGroup, nil
}

func (r *RedisClient) Close() error {
	err := r.client.Close()
	if err != nil {
		r.log.Error("Failed to close redis client", "error", err)
		return fmt.Errorf("failed to close redis client: %v", err)
	}

	r.log.Info("Redis client closed successfully")
	return nil
}
