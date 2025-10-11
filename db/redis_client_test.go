package db

import (
	"context"
	"ermes-central-repo/config"
	"net/url"
	"testing"

	"github.com/ermes-thesis/ermes-common/logger"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"
)

// ↓↓↓ UNCOMMENT THIS IF YOU NEED TO EXECUTE THE TESTS ON A CONTAINER ↓↓↓

//// IsWSL returns true if the Go code is running inside WSL2.
//func IsWSL() bool {
//	if runtime.GOOS != "linux" {
//		return false
//	}
//
//	// Check /proc/version
//	if content, err := os.ReadFile("/proc/version"); err == nil {
//		if strings.Contains(strings.ToLower(string(content)), "microsoft") {
//			return true
//		}
//	}
//
//	// Optionally check /proc/sys/kernel/osrelease (some WSL versions)
//	if content, err := os.ReadFile("/proc/sys/kernel/osrelease"); err == nil {
//		if strings.Contains(strings.ToLower(string(content)), "microsoft") {
//			return true
//		}
//	}
//
//	return false
//}
//
//func GetDockerHost() string {
//	addrs, err := net.LookupHost("host.docker.internal")
//	if err == nil && len(addrs) > 0 {
//		return "host.docker.internal"
//	}
//
//	return "172.17.0.1"
//}

func createNewLogger(t *testing.T) *logger.Logger {
	log, err := logger.NewLogger(logger.Config{
		Level:     zerolog.DebugLevel,
		Format:    "text",
		AddSource: false,
	})
	require.NoError(t, err)
	return log
}

func StartRedisTestContainer(t *testing.T) (*RedisClient, func()) {
	//if !IsWSL() {
	//	t.Setenv("TESTCONTAINERS_HOST_OVERRIDE", GetDockerHost())
	//}
	ctx := context.Background()
	redisContainer, err := tcredis.Run(ctx,
		"redis/redis-stack:latest",
		tcredis.WithSnapshotting(10, 1),
		tcredis.WithLogLevel(tcredis.LogLevelVerbose),
		testcontainers.WithCmd("redis-stack-server"),
		testcontainers.WithCmdArgs("--maxmemory", "500mb"),
		//tcredis.WithConfigFile(filepath.Join("testdata", "redis7.conf")),
	)

	if err != nil {
		t.Fatalf("failed to start container: %s", err)
		return nil, nil
	}

	connectionString, _ := redisContainer.ConnectionString(ctx)
	parsedUrl, _ := url.Parse(connectionString)

	log := createNewLogger(t)
	client, err := NewRedisClient(&config.Config{
		RedisAddr:     parsedUrl.Hostname(),
		RedisPort:     parsedUrl.Port(),
		RedisPassword: "",
		RedisDB:       0,
	}, log)
	require.NoError(t, err)

	return client.(*RedisClient), func() {
		if err := redisContainer.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	}
}

func TestRedisClient(t *testing.T) {
	redisClient, shutdown := StartRedisTestContainer(t)
	defer shutdown()
	defer redisClient.Close()

	// Creating a new group
	err := redisClient.CreateGroup("test_group_name", []string{"participant1", "participant2", "participant3"})
	require.NoError(t, err)

	// Asserting that the new group now exists
	groups, err := redisClient.GetExistingGroups()
	require.NoError(t, err)
	require.Equal(t, 1, len(groups))
	require.Contains(t, groups, "test_group_name")

	// Asserting that all participants are in the group
	groups, err = redisClient.GetUserGroups("participant1")
	require.NoError(t, err)
	require.Equal(t, 1, len(groups))
	require.Contains(t, groups, "test_group_name")
	groups, err = redisClient.GetUserGroups("participant2")
	require.NoError(t, err)
	require.Equal(t, 1, len(groups))
	require.Contains(t, groups, "test_group_name")
	groups, err = redisClient.GetUserGroups("participant2")
	require.NoError(t, err)
	require.Equal(t, 1, len(groups))
	require.Contains(t, groups, "test_group_name")

	// Create an empty group
	err = redisClient.CreateGroup("initial_empty_group", []string{})
	require.NoError(t, err)

	// Check if the group exists and it is really empty
	groups, err = redisClient.GetExistingGroups()
	require.NoError(t, err)
	require.Equal(t, 2, len(groups))
	require.Contains(t, groups, "initial_empty_group")
	require.Contains(t, groups, "test_group_name")

	// Add two participant to the group
	err = redisClient.AddGroupToUser("initial_empty_group", "participant1")
	require.NoError(t, err)
	err = redisClient.AddGroupToUser("initial_empty_group", "participant2")
	require.NoError(t, err)

	// Assert that the two participant were actually added
	groups, err = redisClient.GetUserGroups("participant1")
	require.NoError(t, err)
	require.Equal(t, 2, len(groups))
	require.Contains(t, groups, "initial_empty_group")
	groups, err = redisClient.GetUserGroups("participant2")
	require.NoError(t, err)
	require.Equal(t, 2, len(groups))
	require.Contains(t, groups, "initial_empty_group")

	// Remove a user from a group
	err = redisClient.RemoveGroupFromUser("initial_empty_group", "participant1")
	require.NoError(t, err)

	// Check that the participant has been removed
	groups, err = redisClient.GetUserGroups("participant1")
	require.NoError(t, err)
	require.Equal(t, 1, len(groups))
	require.Contains(t, groups, "test_group_name")
	require.NotContains(t, groups, "initial_empty_group")

	groups, err = redisClient.GetUserGroups("participant2")
	require.NoError(t, err)
	require.Equal(t, 2, len(groups))
	require.Contains(t, groups, "test_group_name")
	require.Contains(t, groups, "initial_empty_group")

	// Delete a group
	groupsBefore, err := redisClient.GetExistingGroups()
	require.NoError(t, err)
	require.Equal(t, 2, len(groupsBefore))
	require.Contains(t, groups, "initial_empty_group")
	require.Contains(t, groups, "test_group_name")
	part1groupsBefore, err := redisClient.GetUserGroups("participant1")
	require.NoError(t, err)
	require.Equal(t, 1, len(part1groupsBefore))
	require.Contains(t, part1groupsBefore, "test_group_name")
	require.NotContains(t, part1groupsBefore, "initial_empty_group")
	err = redisClient.DeleteGroup("initial_empty_group")
	require.NoError(t, err)

	// Assert that the group has been deleted, and no other participant in that group is associated with it
	groupsNow, err := redisClient.GetExistingGroups()
	require.NoError(t, err)
	require.Equal(t, 1, len(groupsNow))
	require.NotContains(t, groupsNow, "initial_empty_group")
	part2groupsNow, err := redisClient.GetUserGroups("participant1")
	require.NoError(t, err)
	require.Equal(t, 1, len(part2groupsNow))
	require.Contains(t, part2groupsNow, "test_group_name")
	require.NotContains(t, part2groupsNow, "initial_empty_group")
}
