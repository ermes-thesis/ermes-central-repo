package db

import (
	"context"
	"ermes-central-repo/config"

	"github.com/ermes-thesis/ermes-common/logger"
	"github.com/redis/go-redis/v9"
)

// GenericDB defines a minimal database abstraction for managing user-group relationships.
// It supports adding/removing users from groups, listing groups and members.
type GenericDB interface {
	// AddGroupToUser associates a user with a group if the group exists.
	//
	// Parameters:
	//	 - groupName: name of the group to add the user to
	//	 - userId: unique identifier of the user
	//
	// Returns:
	//	 - error: non-nil if the group does not exist or the operation fails
	AddGroupToUser(groupName string, userId string) error

	// RemoveGroupFromUser removes a user from a specific group.
	//
	// Parameters:
	//	 - groupName: name of the group to remove the user from
	//	 - userId: unique identifier of the user
	//
	// Returns:
	//	 - error: non-nil if the operation fails
	RemoveGroupFromUser(groupName string, userId string) error

	// GetUserGroups retrieves all groups associated with a user.
	//
	// Parameters:
	//	 - userId: unique identifier of the user
	//
	// Returns:
	//	 - []string: list of group names the user belongs to
	//	 - error: non-nil if retrieval fails
	GetUserGroups(userId string) ([]string, error)

	// CreateGroup creates a new group and optionally assigns initial participants.
	//
	// Parameters:
	//	 - groupName: name of the group to create
	//	 - participantIDs: list of user IDs to include in the group
	//
	// Returns:
	//	 - error: non-nil if the group already exists or creation fails
	CreateGroup(groupName string, participantIDs []string) error

	// DeleteGroup removes a group and its user associations.
	//
	// Parameters:
	//	 - groupName: name of the group to delete
	//
	// Returns:
	//	 - error: non-nil if the group does not exist or deletion fails
	DeleteGroup(groupName string) error

	// GetExistingGroups returns the names of all currently existing groups.
	//
	// Returns:
	//   - []string: list of existing group names
	//   - error: non-nil if retrieval fails
	GetExistingGroups() ([]string, error)

	// GetGroupUsers retrieves all users currently associated with a group.
	//
	// Parameters:
	//   - groupName: name of the group
	//
	// Returns:
	//   - []string: list of user IDs in the group
	//   - error: non-nil if retrieval fails
	GetGroupUsers(groupName string) ([]string, error)

	// Close releases any open database connections.
	//
	// Returns:
	//   - error: non-nil if closing fails
	Close() error
}

// NewRedisClient creates a new Redis-backed implementation of GenericDB.
//
// It establishes a connection to Redis using the configuration parameters and
// verifies connectivity with a ping.
//
// Parameters:
//   - cfg: application configuration containing Redis connection details
//   - log: logger used for logging successes and failures of the implementation methods
//
// Returns:
//   - GenericDB: an initialized Redis client implementing the GenericDB interface
//   - error: non-nil if the connection could not be established
func NewRedisClient(cfg *config.Config, log logger.Logger) (GenericDB, error) {
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
