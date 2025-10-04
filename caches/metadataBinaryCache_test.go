package caches

import (
	"ermes-central-repo/caches/metadata"
	"ermes-central-repo/config"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/ermes-thesis/ermes-common/logger"
	"github.com/rs/zerolog"

	"github.com/stretchr/testify/require"
)

func getTestLogger(t *testing.T) *logger.Logger {
	tmpDir := t.TempDir()
	log, err := logger.NewLogger(logger.Config{
		Level:  zerolog.DebugLevel,
		Format: "text",
		Output: &logger.LogFile{
			LogDir:      tmpDir,
			LogFileName: "test",
		},
		StdoutEnabled: false,
		AddSource:     false,
	})
	require.NoError(t, err)
	return log
}

func TestBinaryCache_SetAndGet(t *testing.T) {
	tmpDir := t.TempDir()
	localizedTmpDir := filepath.ToSlash(tmpDir)
	cfgPath := filepath.Join(tmpDir, "config.yaml")
	err := os.WriteFile(cfgPath, []byte(
		"JWT_EXPIRATION_TIME: \"3600s\"\n"+
			"JWT_PEM_PATH: \"../.devcontainer/private.pem\"\n"+
			"CLIENT_SECRET: \"very VERY secret\"\n"+
			"ERMES_SECRET: \"very secret\"\n"+
			"DEV_SECRET: \"super secret\"\n"+
			"FUNCTION_CACHE_SIZE: \"20\"\n"+
			"FUNCTION_CACHE_TTL: \"1s\"\n"+
			"FUNCTION_CACHE_BIN_DIR: \""+localizedTmpDir+"\"\n"), 0644)
	require.NoError(t, err)

	log := getTestLogger(t)
	defer log.Close()
	loadConfig, err := config.LoadConfig(log, localizedTmpDir)
	require.NoError(t, err)

	cache, err := NewMetadataBinaryCache(loadConfig)
	require.NoError(t, err)

	data := []byte("test-content-binary")
	fn := "func1"
	fnVersion := "1.0.0"

	// Set should write file and cache entry
	meta := metadata.FunctionMetadata{
		Name:                  fn,
		Version:               fnVersion,
		TemplateNames:         nil,
		DeployedOn:            nil,
		StrongCollWriteAccess: false,
	}
	err = cache.Set(meta, data)
	require.NoError(t, err)

	// Check file exists
	binPath := filepath.Join(tmpDir, fn+"@"+fnVersion+".wasm")
	if _, err := os.Stat(binPath); os.IsNotExist(err) {
		require.Fail(t, "binary does not exist")
	}

	// Get should return the same binary content
	m, ok := cache.Get(fn, fnVersion)
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, meta, m)
}

func TestBinaryCache_EvictionRemovesFile(t *testing.T) {
	tmpDir := t.TempDir()
	localizedTmpDir := filepath.ToSlash(tmpDir)
	cfgPath := filepath.Join(tmpDir, "config.yaml")
	err := os.WriteFile(cfgPath, []byte(
		"JWT_EXPIRATION_TIME: \"3600s\"\n"+
			"JWT_PEM_PATH: \"../.devcontainer/private.pem\"\n"+
			"CLIENT_SECRET: \"very VERY secret\"\n"+
			"ERMES_SECRET: \"very secret\"\n"+
			"DEV_SECRET: \"super secret\"\n"+
			"FUNCTION_CACHE_SIZE: \"1\"\n"+
			"FUNCTION_CACHE_TTL: \"0\"\n"+
			"FUNCTION_CACHE_BIN_DIR: \""+localizedTmpDir+"\"\n"), 0644)
	require.NoError(t, err)

	log := getTestLogger(t)
	defer log.Close()
	loadConfig, err := config.LoadConfig(log, localizedTmpDir)
	require.NoError(t, err)

	cache, err := NewMetadataBinaryCache(loadConfig)
	require.NoError(t, err)

	fn1 := "func1"
	fn2 := "func2"
	fnVersion1 := "1.0.0"
	fnVersion2 := "1.0.0"

	// Set should write file and cache entry
	meta := metadata.FunctionMetadata{
		Name:                  fn1,
		Version:               fnVersion1,
		TemplateNames:         nil,
		DeployedOn:            nil,
		StrongCollWriteAccess: false,
	}
	meta2 := metadata.FunctionMetadata{
		Name:                  fn2,
		Version:               fnVersion2,
		TemplateNames:         nil,
		DeployedOn:            nil,
		StrongCollWriteAccess: false,
	}

	data1 := []byte("aaa")
	data2 := []byte("bbb")

	err = cache.Set(meta, data1)
	require.NoError(t, err)
	err = cache.Set(meta2, data2)
	require.NoError(t, err)

	// fn1 should have been evicted, so file removed
	binPath1 := filepath.Join(tmpDir, fn1)
	if _, err := os.Stat(binPath1); !os.IsNotExist(err) {
		require.Fail(t, "binary1 exists but should have been removed")
	}

	// fn2 should still exist
	binPath2 := filepath.Join(tmpDir, fn2+"@"+fnVersion2+".wasm")
	if _, err := os.Stat(binPath2); os.IsNotExist(err) {
		require.Fail(t, "binary2 should still exists but was removed")
	}
}

func TestBinaryCache_GetMissing(t *testing.T) {
	tmpDir := t.TempDir()
	localizedTmpDir := filepath.ToSlash(tmpDir)
	cfgPath := filepath.Join(tmpDir, "config.yaml")
	err := os.WriteFile(cfgPath, []byte(
		"JWT_EXPIRATION_TIME: \"3600s\"\n"+
			"JWT_PEM_PATH: \"../.devcontainer/private.pem\"\n"+
			"CLIENT_SECRET: \"very VERY secret\"\n"+
			"ERMES_SECRET: \"very secret\"\n"+
			"DEV_SECRET: \"super secret\"\n"+
			"FUNCTION_CACHE_SIZE: \"20\"\n"+
			"FUNCTION_CACHE_TTL: \""+strconv.FormatInt((1*time.Second).Nanoseconds(), 10)+"s\"\n"+
			"FUNCTION_CACHE_BIN_DIR: \""+localizedTmpDir+"\"\n"), 0644)
	require.NoError(t, err)

	log := getTestLogger(t)
	defer log.Close()
	loadConfig, err := config.LoadConfig(log, localizedTmpDir)
	require.NoError(t, err)

	cache, err := NewMetadataBinaryCache(loadConfig)
	require.NoError(t, err)

	// Nothing set yet
	m, ok := cache.Get("missing", "missing")
	require.NoError(t, err)
	require.False(t, ok)
	require.Zero(t, m)
}
