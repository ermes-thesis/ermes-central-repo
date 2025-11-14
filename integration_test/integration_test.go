package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"ermes-central-repo/caches"
	"ermes-central-repo/config"
	"ermes-central-repo/db"
	"ermes-central-repo/gateway"
	"ermes-central-repo/tokenEncoder"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"

	common_caches "github.com/ermes-thesis/ermes-common/caches"
	"github.com/ermes-thesis/ermes-common/model"
	"github.com/ermes-thesis/ermes-common/wasm-pool"
	"github.com/golang-jwt/jwt/v5"
	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"

	"github.com/ermes-thesis/ermes-common/logger"
	"github.com/rs/zerolog"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
)

func createNewLogger(t *testing.T) logger.Logger {
	log, err := logger.NewLogger(logger.Config{
		Level:     zerolog.DebugLevel,
		AddSource: true,
		Rolling:   false,
	})
	require.NoError(t, err)
	return log
}

func TestGetGroupToken(t *testing.T) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	t.Cleanup(cancel)

	redisContainer, err := tcredis.Run(ctx,
		"redis/redis-stack:latest",
		tcredis.WithSnapshotting(10, 1),
		tcredis.WithLogLevel(tcredis.LogLevelVerbose),
		testcontainers.WithCmd("redis-stack-server"),
		//testcontainers.WithCmdArgs("--maxmemory", "500mb"),
		//tcredis.WithConfigFile(filepath.Join("testdata", "redis7.conf")),
	)
	defer func() {
		if err := testcontainers.TerminateContainer(redisContainer); err != nil {
			assert.FailNow(t, "Failed to terminate container", err.Error())
		}
	}()
	if err != nil {
		assert.FailNow(t, "Failed to start Redis container", err.Error())
		return
	}

	fmt.Printf("Redis container started with ID: %s\n", redisContainer.GetContainerID())
	endpoint, err := redisContainer.ConnectionString(ctx)
	assert.NoError(t, err, "Failed to get connection string")
	fmt.Printf("Redis endpoint: %s\n", endpoint)

	parsedUrl, _ := url.Parse(endpoint)
	tmpDir := t.TempDir()
	fnRegTestServerMux := http.NewServeMux()
	fnRegTestServerMux.Handle("/retrieve-auth-function/test_auth/1.0.0", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		file, err := os.ReadFile("./authFunction@1.0.0.wasm")
		require.NoError(t, err)

		metadata := model.FunctionMetadata{
			Name:                  "test_auth",
			Version:               "1.0.0",
			TemplateNames:         nil,
			DeployedOn:            nil,
			StrongCollWriteAccess: false,
			IsAuth:                true,
		}
		w.Header().Set("ETag", metadata.Version)
		w.Header().Set("Content-Type", "application/json")
		response := caches.FunctionBundle{
			Binary:   file,
			Metadata: metadata,
		}

		err = json.NewEncoder(w).Encode(response)
		require.NoError(t, err)
		return
	}))
	fnRegTestServer := httptest.NewServer(fnRegTestServerMux)
	fnRegTestServerParsedUrl, err := url.Parse(fnRegTestServer.URL)
	require.NoError(t, err)

	authFn := "test_auth@1.0.0"
	file, err := os.ReadFile("./authFunction@1.0.0.wasm")
	require.NoError(t, err)
	err = os.WriteFile(tmpDir+"/"+authFn, file, 0644)
	require.NoError(t, err)

	cfg := &config.Config{
		JWTPemPath:        "../keys/private.pem",
		JWTExpirationTime: time.Hour * 24,
		RedisAddr:         parsedUrl.Hostname(),
		RedisPort:         parsedUrl.Port(),
		RedisPassword:     "",
		RedisDB:           0,
		WasmRuntime:       wasm.Wasmtime,
		WasmPoolConfig: wasm.Config{
			PoolSize:  10,
			Timeout:   time.Second * 5,
			WithCache: true,
			IsAot:     false,
		},
		ErmesSecret:             "secret",
		ClientSecret:            "client_secret",
		DevSecret:               "dev_secret",
		FunctionRegistryAddress: fnRegTestServerParsedUrl.Hostname(),
		FunctionRegistryPort:    fnRegTestServerParsedUrl.Port(),
		FunctionCache: common_caches.MetadataBinaryCacheConfig{
			Size:   100,
			Ttl:    0,
			BinDir: tmpDir,
		},
		LatestCache: common_caches.LatestCacheConfig{
			Size: 100,
			Ttl:  time.Hour * 12,
		},
	}

	log := createNewLogger(t)
	redisDb, err := db.NewRedisClient(cfg, log)
	require.NoError(t, err)
	err = redisDb.CreateGroup("test_group", []string{"userId"})
	require.NoError(t, err)
	err = redisDb.CreateGroup("another_test_group", []string{"userId"})
	require.NoError(t, err)
	err = redisDb.CreateGroup("third_group", []string{})
	require.NoError(t, err)
	err = redisDb.AddGroupToUser("third_group", "userId")

	wasmPool := wasm.NewWasmedgeExecutorPool(cfg.WasmPoolConfig, createNewLogger(t))
	tknEncoder, err := tokenEncoder.NewGroupTokenEncoder(cfg.JWTPemPath, cfg.JWTExpirationTime)
	require.NoError(t, err)
	fnCache, err := caches.NewDefaultFunctionCache(cfg, log)
	require.NoError(t, err)

	app := gateway.NewServer(cfg, log, wasmPool, redisDb, tknEncoder, fnCache, "wrongFn@1.0.0")
	srv := httptest.NewServer(app.Router())
	defer srv.Close()

	// Set auth function
	req, err := http.NewRequest(http.MethodPut, srv.URL+"/set-auth-function/test_auth/1.0.0", nil)
	assert.NoError(t, err, "Failed to create request")

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer token")
	req.Header.Set("X-DEV-SECRET", "dev_secret")

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err, "Failed to send request")
	err = res.Body.Close()
	require.NoError(t, err)

	req, err = http.NewRequest(http.MethodGet, srv.URL+"/token", nil)
	assert.NoError(t, err, "Failed to create request")

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer token")
	req.Header.Set("X-CLIENT-SECRET", "client_secret")

	res, err = http.DefaultClient.Do(req)
	assert.NoError(t, err, "Failed to send request")
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	var response map[string]string
	err = json.Unmarshal(body, &response)
	require.NoError(t, err)

	token, err := jwt.ParseWithClaims(response["group_token"], &tokenEncoder.GroupTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		pKeyBytes, err := os.ReadFile("../keys/public.pem")
		if err != nil {
			return nil, err
		}
		return jwt.ParseRSAPublicKeyFromPEM(pKeyBytes)
	})
	require.NoError(t, err)
	require.True(t, token.Valid)
	customClaims, ok := token.Claims.(*tokenEncoder.GroupTokenClaims)
	require.True(t, ok)

	require.Equal(t, "userId", customClaims.Subject)
	require.Equal(t, []string{"test_group", "another_test_group", "third_group"}, customClaims.Groups)
	require.Equal(t, "ermes-central-repository", customClaims.Issuer)
	require.True(t, customClaims.ExpiresAt.After(time.Now()))

	require.Equalf(t, http.StatusOK, res.StatusCode, "Expected status code 200 OK, got %d: %s", res.StatusCode, string(body))
}

func TestGroupChangesInteraction(t *testing.T) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	t.Cleanup(cancel)

	redisContainer, err := tcredis.Run(ctx,
		"redis/redis-stack:latest",
		tcredis.WithSnapshotting(10, 1),
		tcredis.WithLogLevel(tcredis.LogLevelVerbose),
		testcontainers.WithCmd("redis-stack-server"),
		testcontainers.WithCmdArgs("--maxmemory", "500mb"),
		//tcredis.WithConfigFile(filepath.Join("testdata", "redis7.conf")),
	)
	defer func() {
		if err := testcontainers.TerminateContainer(redisContainer); err != nil {
			assert.FailNow(t, "Failed to terminate container", err.Error())
		}
	}()
	if err != nil {
		assert.FailNow(t, "Failed to start Redis container", err.Error())
		return
	}

	fmt.Printf("Redis container started with ID: %s\n", redisContainer.GetContainerID())
	endpoint, err := redisContainer.ConnectionString(ctx)
	assert.NoError(t, err, "Failed to get connection string")
	fmt.Printf("Redis endpoint: %s\n", endpoint)

	parsedUrl, _ := url.Parse(endpoint)
	tmpDir := t.TempDir()

	authFn := "test_auth@1.0.0"

	cfg := &config.Config{
		JWTPemPath:        "../keys/private.pem",
		JWTExpirationTime: time.Hour * 24,
		RedisAddr:         parsedUrl.Hostname(),
		RedisPort:         parsedUrl.Port(),
		RedisPassword:     "",
		RedisDB:           0,
		WasmRuntime:       wasm.Wasmtime,
		WasmPoolConfig: wasm.Config{
			PoolSize:  10,
			Timeout:   time.Second * 5,
			WithCache: true,
			IsAot:     false,
		},
		ErmesSecret:  "secret",
		ClientSecret: "client_secret",
		DevSecret:    "dev_secret",
		FunctionCache: common_caches.MetadataBinaryCacheConfig{
			Size:   100,
			Ttl:    0,
			BinDir: tmpDir,
		},
		LatestCache: common_caches.LatestCacheConfig{
			Size: 100,
			Ttl:  time.Hour * 12,
		},
	}

	log := createNewLogger(t)
	redisDb, err := db.NewRedisClient(cfg, log)
	require.NoError(t, err)
	err = redisDb.CreateGroup("test_group", []string{"userId"})
	require.NoError(t, err)
	err = redisDb.CreateGroup("another_test_group", []string{"userId"})
	require.NoError(t, err)
	err = redisDb.CreateGroup("third_group", []string{})
	require.NoError(t, err)
	err = redisDb.AddGroupToUser("third_group", "userId")

	wasmPool := wasm.NewWasmedgeExecutorPool(cfg.WasmPoolConfig, createNewLogger(t))
	tknEncoder, err := tokenEncoder.NewGroupTokenEncoder(cfg.JWTPemPath, cfg.JWTExpirationTime)
	require.NoError(t, err)
	fnCache, err := caches.NewDefaultFunctionCache(cfg, log)
	require.NoError(t, err)

	app := gateway.NewServer(cfg, log, wasmPool, redisDb, tknEncoder, fnCache, authFn)
	srv := httptest.NewServer(app.Router())
	defer srv.Close()

	// Receive error since we tried to add a user to a non-existing group
	req, err := http.NewRequest(http.MethodPost, srv.URL+"/group/first_group/user1", nil)
	assert.NoError(t, err, "Failed to create request")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer token")
	req.Header.Set("X-ERMES-SECRET", "secret")

	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	b, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	require.Equal(t, http.StatusInternalServerError, res.StatusCode)
	require.Equal(t, "Internal Server Error: could not add user user1 to group first_group: failed to add user to group: group does not exist\n", string(b))

	// StatusCreated --> Group created
	req, err = http.NewRequest(http.MethodPost, srv.URL+"/group/first_group", nil)
	assert.NoError(t, err, "Failed to create request")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer token")
	req.Header.Set("X-ERMES-SECRET", "secret")

	res, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, res.StatusCode)

	// StatusOK --> User1 added to group
	req, err = http.NewRequest(http.MethodPost, srv.URL+"/group/first_group/user1", nil)
	assert.NoError(t, err, "Failed to create request")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer token")
	req.Header.Set("X-ERMES-SECRET", "secret")

	res, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.StatusCode)

	// StatusOK --> Second group created with user1 and user2
	participantIDs := []string{"user1", "user2"}
	b, err = json.Marshal(participantIDs)
	req, err = http.NewRequest(http.MethodPost, srv.URL+"/group/second_group", bytes.NewBuffer(b))
	assert.NoError(t, err, "Failed to create request")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer token")
	req.Header.Set("X-ERMES-SECRET", "secret")

	res, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, res.StatusCode)

	// StatusOK --> User2 removed from second group
	req, err = http.NewRequest(http.MethodDelete, srv.URL+"/group/second_group/user2", nil)
	assert.NoError(t, err, "Failed to create request")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer token")
	req.Header.Set("X-ERMES-SECRET", "secret")

	res, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.StatusCode)

	// StatusOK --> First group deleted
	req, err = http.NewRequest(http.MethodDelete, srv.URL+"/group/first_group", nil)
	assert.NoError(t, err, "Failed to create request")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer token")
	req.Header.Set("X-ERMES-SECRET", "secret")

	res, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.StatusCode)
}
