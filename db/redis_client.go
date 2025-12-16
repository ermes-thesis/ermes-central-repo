package db

import (
	"context"
	"errors"
	"fmt"
	"github.com/ermes-thesis/ermes-common/logger"
	"github.com/redis/go-redis/v9"
	"math/rand"
	"strings"
	"time"
)

// RedisClient implements the GenericDB interface using Redis as the storage backend.
// It ensures atomicity using Redis transactions and exponential backoff for retries.
type RedisClient struct {
	client *redis.Client
	log    logger.Logger
}

const maxRetries = 10
const defaultBackoff = time.Millisecond * 10
const maxBackoff = time.Second * 1

// AddGroupToUser adds a user to a group if the group exists. Retries on concurrent updates.
func (r *RedisClient) AddGroupToUser(groupName string, userId string) error {
	ctx := context.Background()
	backoff := defaultBackoff

	userKey := buildUserGroupsKey(userId)
	groupExistenceKey := buildExistenceGroupKey(groupName)
	groupKey := buildGroupParticipantsKey(groupName)

	for retries := 0; retries < maxRetries; retries++ {
		err := r.client.Watch(ctx, func(tx *redis.Tx) error {
			exists, err := tx.Exists(ctx, groupExistenceKey).Result()
			if err != nil {
				r.log.Error("Failed to check if group exists", "group_name", groupName, "user_id", userId, "error", err)
				return fmt.Errorf("failed to check if group exists: %v", err)
			}
			if exists == 0 {
				r.log.Warn("Attempted to add user to non-existing group", "group_name", groupName, "user_id", userId)
				return errors.New("group does not exist")
			}

			_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
				pipe.SAdd(ctx, userKey, groupName)
				pipe.SAdd(ctx, groupKey, userId)
				return nil
			})
			if err != nil {
				r.log.Error("Failed to add user to group", "group_name", groupName, "user_id", userId, "error", err)
				return fmt.Errorf("failed to add user to group: %v", err)
			}
			return nil
		}, groupExistenceKey)

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

// RemoveGroupFromUser removes a user from a group.
func (r *RedisClient) RemoveGroupFromUser(groupName string, userId string) error {
	ctx := context.Background()
	userKey := buildUserGroupsKey(userId)
	groupKey := buildGroupParticipantsKey(groupName)

	_, err := r.client.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.SRem(ctx, userKey, groupName)
		pipe.SRem(ctx, groupKey, userId)

		return nil
	})

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

// GetUserGroups retrieves the list of groups a user belongs to.
func (r *RedisClient) GetUserGroups(userId string) ([]string, error) {
	ctx := context.Background()

	userKey := buildUserGroupsKey(userId)
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

// CreateGroup creates a new group and associates the given participants.
func (r *RedisClient) CreateGroup(groupName string, participantIDs []string) error {
	ctx := context.Background()

	backoff := defaultBackoff

	groupExistenceKey := buildExistenceGroupKey(groupName)
	groupKey := buildGroupParticipantsKey(groupName)

	var cmds []redis.Cmder
	for retries := 0; retries < maxRetries; retries++ {
		err := r.client.Watch(ctx, func(tx *redis.Tx) error {
			exists, err := tx.Exists(ctx, groupExistenceKey).Result()
			if err != nil {
				return err
			}
			if exists == 1 {
				return errors.New("group already exists")
			}

			cmds, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
				// Create special record that indicate the group existence
				pipe.Set(ctx, groupExistenceKey, "", 0)

				for _, participantID := range participantIDs {
					pipe.SAdd(ctx, buildUserGroupsKey(participantID), groupName)
					pipe.SAdd(ctx, groupKey, participantID)
				}
				return nil
			})

			return err
		}, groupKey)

		if errors.Is(err, redis.TxFailedErr) {
			jitter := time.Duration(rand.Int63n(int64(backoff / 2)))
			time.Sleep(backoff + jitter)

			backoff *= 2
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
			r.log.Warn("Failed to create group due to a concurrent creation/deletion.", "group_name", groupName, "participants", participantIDs, "retry_num", retries, "max_retries", maxRetries, "backing_off_for", backoff)
			continue
		}

		if err != nil {
			r.log.Error("Transaction failed",
				"group_name", groupName,
				"participant_ids", participantIDs,
				"error", err,
			)

			if r.log.Enabled(logger.DebugLevel) {
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

	// We arrive here only if all the retries have failed
	r.log.Error("Transaction retries exhausted for AddGroupToUser",
		"group_name", groupName,
		"participantIDs", participantIDs,
	)
	return errors.New("transaction retries exhausted")
}

// DeleteGroup removes a group and its user associations.
func (r *RedisClient) DeleteGroup(groupName string) error {
	ctx := context.Background()

	backoff := defaultBackoff

	groupExistenceKey := buildExistenceGroupKey(groupName)
	groupKey := buildGroupParticipantsKey(groupName)

	for retries := 0; retries < maxRetries; retries++ {
		err := r.client.Watch(ctx, func(tx *redis.Tx) error {
			exists, err := tx.Exists(ctx, groupExistenceKey).Result()
			if err != nil {
				return err
			}
			if exists == 0 {
				return errors.New("group does not exist")
			}

			// Get all members atomically within the transaction
			members, err := tx.SMembers(ctx, groupKey).Result()
			if err != nil {
				return fmt.Errorf("failed to get group members: %v", err)
			}

			_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
				// Remove group from all users
				for _, userID := range members {
					pipe.SRem(ctx, buildUserGroupsKey(userID), groupName)
				}
				// Delete the group marker and members set
				pipe.Del(ctx, groupKey)
				pipe.Del(ctx, groupExistenceKey)

				return nil
			})

			return err
		}, groupKey)

		if errors.Is(err, redis.TxFailedErr) {
			jitter := time.Duration(rand.Int63n(int64(backoff / 2)))
			time.Sleep(backoff + jitter)

			backoff *= 2
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
			r.log.Warn("Failed to delete group due to concurrent modification",
				"group_name", groupName,
				"retry_num", retries)
			continue
		}

		if err != nil {
			r.log.Error("Failed to delete group",
				"group_name", groupName,
				"error", err)
			return fmt.Errorf("failed to delete group: %v", err)
		}

		r.log.Infof("Group '%s' deleted successfully", groupName)
		return nil
	}

	return errors.New("transaction retries exhausted")
}

// GetExistingGroups retrieves the list of all existing groups.
func (r *RedisClient) GetExistingGroups() ([]string, error) {
	ctx := context.Background()

	var existingGroups []string
	var cursor uint64
	for {
		var batchGroups []string
		var err error
		batchGroups, cursor, err = r.client.Scan(ctx, cursor, "info:*", 100).Result()
		if err != nil {
			r.log.Error("Failed to scan existing groups",
				"error", err)
			r.log.Debug("Failed to scan existing groups",
				"cursor", cursor,
				"batch_groups", batchGroups,
			)
			return nil, fmt.Errorf("failed to scan existing groups: %v", err)
		}

		for _, groupKey := range batchGroups {
			existingGroups = append(existingGroups, strings.Split(groupKey, ":")[1])
		}

		if cursor == 0 {
			break
		}
	}

	r.log.Infof("Retrieved existing groups successfully: %v", existingGroups[:min(len(existingGroups), 20)])
	r.log.Debugf("Retrieved existing groups successfully: %v", existingGroups)
	return existingGroups, nil
}

// GetGroupUsers retrieves all users associated with a group.
func (r *RedisClient) GetGroupUsers(groupName string) ([]string, error) {
	ctx := context.Background()

	var usersInGroup []string

	groupKey := buildGroupParticipantsKey(groupName)

	usersInGroup, err := r.client.SMembers(ctx, groupKey).Result()
	if err != nil {
		r.log.Error("Failed to get group users",
			"group_name", groupName,
			"error", err)
		return nil, fmt.Errorf("failed to get group users: %v", err)
	}

	r.log.Infof("Retrieved group users successfully: %v", usersInGroup[:min(len(usersInGroup), 20)])
	r.log.Debugf("Retrieved group users successfully: %v", usersInGroup)
	return usersInGroup, nil
}

// Close closes the Redis client connection.
func (r *RedisClient) Close() error {
	err := r.client.Close()
	if err != nil {
		r.log.Error("Failed to close redis client", "error", err)
		return fmt.Errorf("failed to close redis client: %v", err)
	}

	r.log.Info("Redis client closed successfully")
	return nil
}

// Helper key builders
func buildExistenceGroupKey(groupName string) string {
	return fmt.Sprintf("info:%v", groupName)
}
func buildGroupParticipantsKey(groupName string) string {
	return fmt.Sprintf("group:%v", groupName)
}
func buildUserGroupsKey(userId string) string {
	return fmt.Sprintf("user:%v", userId)
}
