package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ermes-thesis/ermes-common/caches"
	"github.com/ermes-thesis/ermes-common/logger"
	"github.com/ermes-thesis/ermes-common/wasm-pool"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig_FromSpecificPath(t *testing.T) {
	log, err := logger.NewLogger(logger.Config{
		Level:         logger.DebugLevel,
		Format:        "text",
		Output:        nil,
		StdoutEnabled: false,
	})
	defer log.Close()
	require.NoError(t, err)

	tmpDir := t.TempDir()
	configFile := []byte(
		"JWT_PEM_PATH: \"keys/private.pem\"\n" +
			"JWT_EXPIRATION_TIME: \"2000s\"\n" +
			"REDIS_ADDR: \"redis\"\n" +
			"REDIS_PORT: \"6379\"\n" +
			"REDIS_PASSWORD: \"eYVX7EwVmmxKPCDmwMtyKVge8oLd2t81\"\n" +
			"REDIS_DB: \"0\"\n" +
			"HTTP_PORT: \"8080\"\n" +
			"WASM_RUNTIME: \"0\"\n" +
			"WASM_POOL_CONFIG: \n" +
			"  poolSize: \"10\"\n" +
			"ERMES_SECRET: \"secret\"\n" +
			"CLIENT_SECRET: \"client_secret\"\n" +
			"DEV_SECRET: \"dev_secret\"\n" +
			"FUNCTION_REGISTRY_ADDRESS: \"host.docker.internal\"")
	err = os.WriteFile(filepath.Join(tmpDir, "config.yaml"), configFile, os.ModePerm)
	require.NoError(t, err)

	cfg, err := LoadConfig(log, tmpDir)
	require.NoError(t, err)

	require.Equal(t, "keys/private.pem", cfg.JWTPemPath)
	require.Equal(t, time.Second*2000, cfg.JWTExpirationTime)
	require.Equal(t, "redis", cfg.RedisAddr)
	require.Equal(t, "6379", cfg.RedisPort)
	require.Equal(t, "eYVX7EwVmmxKPCDmwMtyKVge8oLd2t81", cfg.RedisPassword)
	require.Equal(t, 0, cfg.RedisDB)
	require.Equal(t, "8080", cfg.HTTPPort)
	require.Equal(t, wasm.Wasmedge, cfg.WasmRuntime)
	require.Equal(t, 10, cfg.WasmPoolConfig.PoolSize)
	require.Equal(t, "secret", cfg.ErmesSecret)
	require.Equal(t, "client_secret", cfg.ClientSecret)
	require.Equal(t, "dev_secret", cfg.DevSecret)
	require.Equal(t, "host.docker.internal", cfg.FunctionRegistryAddress)
}

func TestLoadConfig_FromCurrentDirectory(t *testing.T) {
	log, err := logger.NewLogger(logger.Config{
		Level:         logger.DebugLevel,
		Format:        "text",
		Output:        nil,
		StdoutEnabled: false,
	})
	require.NoError(t, err)

	tmpDir := t.TempDir()
	configFile := []byte(
		"JWT_PEM_PATH: \"keys/private.pem\"\n" +
			"JWT_EXPIRATION_TIME: \"2000s\"\n" +
			"REDIS_ADDR: \"redis\"\n" +
			"REDIS_PORT: \"6379\"\n" +
			"REDIS_PASSWORD: \"eYVX7EwVmmxKPCDmwMtyKVge8oLd2t81\"\n" +
			"REDIS_DB: \"0\"\n" +
			"HTTP_PORT: \"8080\"\n" +
			"WASM_RUNTIME: \"0\"\n" +
			"WASM_POOL_CONFIG: \n" +
			"  poolSize: \"10\"\n" +
			"ERMES_SECRET: \"secret\"\n" +
			"CLIENT_SECRET: \"client_secret\"\n" +
			"DEV_SECRET: \"dev_secret\"\n" +
			"FUNCTION_REGISTRY_ADDRESS: \"host.docker.internal\"")
	err = os.WriteFile(filepath.Join(tmpDir, "config.yaml"), configFile, os.ModePerm)
	t.Chdir(tmpDir)
	require.NoError(t, err)

	cfg, err := LoadConfig(log, "")
	require.NoError(t, err)

	require.Equal(t, "keys/private.pem", cfg.JWTPemPath)
	require.Equal(t, time.Second*2000, cfg.JWTExpirationTime)
	require.Equal(t, "redis", cfg.RedisAddr)
	require.Equal(t, "6379", cfg.RedisPort)
	require.Equal(t, "eYVX7EwVmmxKPCDmwMtyKVge8oLd2t81", cfg.RedisPassword)
	require.Equal(t, 0, cfg.RedisDB)
	require.Equal(t, "8080", cfg.HTTPPort)
	require.Equal(t, wasm.Wasmedge, cfg.WasmRuntime)
	require.Equal(t, 10, cfg.WasmPoolConfig.PoolSize)
	require.Equal(t, "secret", cfg.ErmesSecret)
	require.Equal(t, "client_secret", cfg.ClientSecret)
	require.Equal(t, "dev_secret", cfg.DevSecret)
	require.Equal(t, "host.docker.internal", cfg.FunctionRegistryAddress)
}

func TestDefaultValues(t *testing.T) {
	log, err := logger.NewLogger(logger.Config{
		Level:         logger.DebugLevel,
		Format:        "text",
		Output:        nil,
		StdoutEnabled: false,
	})
	require.NoError(t, err)

	tmpDir := t.TempDir()
	configFile := []byte(
		"JWT_PEM_PATH: \"keys/private.pem\"\n" +
			"ERMES_SECRET: \"secret\"\n" +
			"CLIENT_SECRET: \"client_secret\"\n" +
			"DEV_SECRET: \"dev_secret\"\n")

	err = os.WriteFile(filepath.Join(tmpDir, "config.yaml"), configFile, os.ModePerm)
	require.NoError(t, err)

	cfg, err := LoadConfig(log, tmpDir)
	require.NoError(t, err)
	require.Equal(t, "keys/private.pem", cfg.JWTPemPath)
	require.Equal(t, time.Second*3600, cfg.JWTExpirationTime)
	require.Equal(t, "redis", cfg.RedisAddr)
	require.Equal(t, "6379", cfg.RedisPort)
	require.Equal(t, 0, cfg.RedisDB)
	require.Equal(t, "8080", cfg.HTTPPort)
	require.Equal(t, wasm.Wasmtime, cfg.WasmRuntime)
	require.Equal(t, 30, cfg.WasmPoolConfig.PoolSize)
	require.Equal(t, "secret", cfg.ErmesSecret)
	require.Equal(t, "client_secret", cfg.ClientSecret)
	require.Equal(t, "dev_secret", cfg.DevSecret)
	require.Equal(t, "localhost", cfg.FunctionRegistryAddress)
	require.Equal(t, "8080", cfg.FunctionRegistryPort)
	def := caches.MetadataBinaryCacheConfig{
		Size:   100,
		Ttl:    0,
		BinDir: "./function_cache",
	}
	require.Equal(t, def.Size, cfg.FunctionCache.Size)
	require.Equal(t, def.Ttl, cfg.FunctionCache.Ttl)
	require.Equal(t, def.BinDir, cfg.FunctionCache.BinDir)
}

func TestConfigFileNotFound(t *testing.T) {
	log, err := logger.NewLogger(logger.Config{
		Level:         logger.DebugLevel,
		Format:        "text",
		Output:        nil,
		StdoutEnabled: false,
	})
	require.NoError(t, err)

	_, err = LoadConfig(log, "")
	require.ErrorContains(t, err, "config.yaml file not found in any defined paths:")
}

func TestMissingSecretLeadToFail(t *testing.T) {
	log, err := logger.NewLogger(logger.Config{
		Level:         logger.DebugLevel,
		Format:        "text",
		Output:        nil,
		StdoutEnabled: false,
	})
	require.NoError(t, err)

	tmpDir := t.TempDir()
	configFile := []byte("REDIS_PORT: \"6379\"\n")
	err = os.WriteFile(filepath.Join(tmpDir, "config.yaml"), configFile, os.ModePerm)
	require.NoError(t, err)

	_, err = LoadConfig(log, tmpDir)
	require.EqualError(t, err, "missing required config values: [jwt_pem_path ermes_secret client_secret dev_secret]")
}
