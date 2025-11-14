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

func createNewLogger(t *testing.T) logger.Logger {
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

func TestCreateGroup(t *testing.T) {
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
}

func TestAddGroupToUser(t *testing.T) {
	redisClient, shutdown := StartRedisTestContainer(t)
	defer shutdown()
	defer redisClient.Close()

	// Create an empty group
	err := redisClient.CreateGroup("initial_empty_group", []string{})
	require.NoError(t, err)

	// Check if the group exists
	groups, err := redisClient.GetExistingGroups()
	require.NoError(t, err)
	require.Equal(t, 1, len(groups))
	require.Contains(t, groups, "initial_empty_group")

	// Add two participant to the group
	err = redisClient.AddGroupToUser("initial_empty_group", "participant1")
	require.NoError(t, err)
	err = redisClient.AddGroupToUser("initial_empty_group", "participant2")
	require.NoError(t, err)

	// Assert that the two participant were actually added
	groups, err = redisClient.GetUserGroups("participant1")
	require.NoError(t, err)
	require.Equal(t, 1, len(groups))
	require.Contains(t, groups, "initial_empty_group")
	groups, err = redisClient.GetUserGroups("participant2")
	require.NoError(t, err)
	require.Equal(t, 1, len(groups))
	require.Contains(t, groups, "initial_empty_group")
}

func TestAddGroupToUser_Fail(t *testing.T) {
	redisClient, shutdown := StartRedisTestContainer(t)
	defer shutdown()
	defer redisClient.Close()

	// Create an empty group
	err := redisClient.CreateGroup("initial_empty_group", []string{})
	require.NoError(t, err)

	// Check if the group exists
	groups, err := redisClient.GetExistingGroups()
	require.NoError(t, err)
	require.Equal(t, 1, len(groups))
	require.Contains(t, groups, "initial_empty_group")

	// Add two participant to the group
	err = redisClient.AddGroupToUser("non_existing_group", "participant1")
	require.EqualError(t, err, "failed to add user to group: group does not exist")
}

func TestRemoveGroupFromUser(t *testing.T) {
	redisClient, shutdown := StartRedisTestContainer(t)
	defer shutdown()
	defer redisClient.Close()

	err := redisClient.CreateGroup("first_group", []string{"participant1", "participant2", "participant3"})
	require.NoError(t, err)
	err = redisClient.CreateGroup("second_group", []string{"participant1", "participant2"})
	require.NoError(t, err)

	groups, err := redisClient.GetUserGroups("participant1")
	require.NoError(t, err)
	require.Equal(t, 2, len(groups))
	require.Contains(t, groups, "first_group")
	require.Contains(t, groups, "second_group")

	err = redisClient.RemoveGroupFromUser("first_group", "participant1")
	require.NoError(t, err)
	groups, err = redisClient.GetUserGroups("participant1")
	require.NoError(t, err)
	require.Equal(t, 1, len(groups))
	require.NotContains(t, groups, "first_group")
	require.Contains(t, groups, "second_group")

	// Check silent fail
	err = redisClient.RemoveGroupFromUser("first_group", "non_existing_user")
	require.NoError(t, err)
	users, err := redisClient.GetGroupUsers("first_group")
	require.NoError(t, err)
	require.Equal(t, 2, len(users))
	require.Contains(t, users, "participant2")
	require.Contains(t, users, "participant3")
}

func TestRemoveGroupFromUser_Fail(t *testing.T) {
	redisClient, shutdown := StartRedisTestContainer(t)
	defer shutdown()
	defer redisClient.Close()

	err := redisClient.CreateGroup("first_group", []string{"participant1", "participant2", "participant3"})
	require.NoError(t, err)

}

func TestDeleteGroup(t *testing.T) {
	redisClient, shutdown := StartRedisTestContainer(t)
	defer shutdown()
	defer redisClient.Close()

	err := redisClient.CreateGroup("first_group", []string{"participant1", "participant2", "participant3"})
	require.NoError(t, err)
	err = redisClient.CreateGroup("second_group", []string{"participant1", "participant2"})
	require.NoError(t, err)

	groups, err := redisClient.GetExistingGroups()
	require.NoError(t, err)
	require.Equal(t, 2, len(groups))
	require.Contains(t, groups, "first_group")
	require.Contains(t, groups, "second_group")

	err = redisClient.DeleteGroup("first_group")
	require.NoError(t, err)

	groups, err = redisClient.GetExistingGroups()
	require.NoError(t, err)
	require.Equal(t, 1, len(groups))
	require.NotContains(t, groups, "first_group")
	require.Contains(t, groups, "second_group")

	user1groups, err := redisClient.GetUserGroups("participant1")
	require.NoError(t, err)
	require.Equal(t, 1, len(user1groups))
	require.Contains(t, user1groups, "second_group")
	require.NotContains(t, user1groups, "first_group")

	user2groups, err := redisClient.GetUserGroups("participant2")
	require.NoError(t, err)
	require.Equal(t, 1, len(user2groups))
	require.Contains(t, user2groups, "second_group")
	require.NotContains(t, user2groups, "first_group")
}

func TestGetUserGroups(t *testing.T) {
	redisClient, shutdown := StartRedisTestContainer(t)
	defer shutdown()
	defer redisClient.Close()

	err := redisClient.CreateGroup("first_group", []string{"participant1", "participant2", "participant3"})
	require.NoError(t, err)
	err = redisClient.CreateGroup("second_group", []string{"participant1", "participant2"})
	require.NoError(t, err)
	err = redisClient.CreateGroup("third_group", []string{"participant2", "participant4"})

	groups, err := redisClient.GetUserGroups("participant1")
	require.NoError(t, err)
	require.Equal(t, 2, len(groups))
	require.Contains(t, groups, "first_group")
	require.Contains(t, groups, "second_group")
	require.NotContains(t, groups, "third_group")

	groups, err = redisClient.GetUserGroups("non_existing_user")
	require.NoError(t, err)
	require.Empty(t, groups)
}

func TestGetGroupUsers(t *testing.T) {
	redisClient, shutdown := StartRedisTestContainer(t)
	defer shutdown()
	defer redisClient.Close()

	err := redisClient.CreateGroup("first_group", []string{"participant1", "participant2", "participant3"})
	require.NoError(t, err)

	groups, err := redisClient.GetGroupUsers("first_group")
	require.NoError(t, err)
	require.Equal(t, 3, len(groups))
	require.Contains(t, groups, "participant1")
	require.Contains(t, groups, "participant2")
	require.Contains(t, groups, "participant3")
}
