package db

import (
	"context"
	"errors"
	"fmt"
	"github.com/ermes-thesis/ermes-common/logger"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"math/rand"
	"strings"
	"time"
)

type RedisClient struct {
	client *redis.Client
	log    *logger.Logger
}

const maxRetries = 10
const defaultBackoff = time.Millisecond * 10
const maxBackoff = time.Second * 1

func (r *RedisClient) AddGroupToUser(groupName string, userId string) error {
	ctx := context.Background()
	backoff := defaultBackoff

	userKey := "data:" + userId

	for retries := 0; retries < maxRetries; retries++ {
		err := r.client.Watch(ctx, func(tx *redis.Tx) error {
			exists, err := tx.SIsMember(ctx, GroupListKey, groupName).Result()
			if err != nil {
				r.log.Error("Failed to check if group exists", "group_name", groupName, "user_id", userId, "error", err)
				return fmt.Errorf("failed to check if group exists: %v", err)
			}
			if !exists {
				r.log.Warn("Attempted to add user to non-existing group", "group_name", groupName, "user_id", userId)
				return errors.New("group does not exist")
			}

			_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
				pipe.SAdd(ctx, userKey, groupName)
				return nil
			})
			if err != nil {
				r.log.Error("Failed to add user to group", "group_name", groupName, "user_id", userId, "error", err)
				return fmt.Errorf("failed to add user to group: %v", err)
			}
			return nil
		}, GroupListKey)

		if errors.Is(err, redis.TxFailedErr) {
			jitter := time.Duration(rand.Int63n(int64(backoff / 2)))
			time.Sleep(backoff + jitter)

			backoff *= 2
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
			r.log.Warn("Failed to add user to group due to a concurrent update.", "group_name", groupName, "user_id", userId, "retry_num", retries, "max_retries", maxRetries, "backing_off_for", backoff)
			continue
		}

		// If it is any other Redis error or if it is Nil (success) return
		if err != nil {
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

	// We arrive here only if all the retries have failed
	r.log.Error("Transaction retries exhausted for AddGroupToUser",
		"group_name", groupName,
		"user_id", userId,
	)
	return errors.New("transaction retries exhausted")
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

	cmds, err := r.client.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.SAdd(ctx, GroupListKey, groupName)

		for _, participantID := range participantIDs {
			pipe.SAdd(ctx, "data:"+participantID, groupName)
		}
		return nil
	})

	if err != nil {
		r.log.Error("Transaction failed",
			"group_name", groupName,
			"participant_ids", participantIDs,
			"error", err,
		)

		if r.log.Enabled(zerolog.DebugLevel) {
			for i, cmd := range cmds {
				if err := cmd.Err(); err != nil {
					if i == 0 {
						r.log.Debug("Failed to create group",
							"group_name", groupName,
							"participant_ids", participantIDs,
							"error", err,
						)
					} else {
						r.log.Debug("Failed to add participant to group",
							"group_name", groupName,
							"participant_id", participantIDs[i+1],
							"error", err,
						)
					}
				}
			}
		}

		return fmt.Errorf("failed to create group: %v", err)
	}

	r.log.Info("Group created successfully",
		"group_name", groupName,
		"participant_ids", participantIDs,
	)
	return nil
}

// 	TODO: Atm this cannot be done atomically, meaning that there can be inconsistencies. To do this atomically we would need a reverse index
// 		  storing for each group its members or add some locks

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
