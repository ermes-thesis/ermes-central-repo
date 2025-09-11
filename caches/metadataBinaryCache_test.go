package caches

import (
	"caches/metadata"
	"ermes-central-repo/config"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestBinaryCache_SetAndGet(t *testing.T) {
	t.Setenv("JWT_EXPIRATION_TIME", "3600")
	t.Setenv("JWT_PEM_PATH", "../.devcontainer/private.pem")
	t.Setenv("MIDDLEWARE_SECRET", "very VERY secret")
	t.Setenv("ERMES_SECRET", "very secret")
	t.Setenv("DEV_SECRET", "super secret")

	t.Setenv("FUNCTION_CACHE_SIZE", "20")
	t.Setenv("FUNCTION_CACHE_TTL", strconv.FormatInt((1*time.Second).Nanoseconds(), 10))
	tmpDir := t.TempDir()
	t.Setenv("FUNCTION_CACHE_BIN_DIR", tmpDir)
	config.LoadConfig()

	cache, err := NewMetadataBinaryCache()
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
	t.Setenv("JWT_EXPIRATION_TIME", "3600")
	t.Setenv("JWT_PEM_PATH", "../.devcontainer/private.pem")
	t.Setenv("MIDDLEWARE_SECRET", "very VERY secret")
	t.Setenv("ERMES_SECRET", "very secret")
	t.Setenv("DEV_SECRET", "super secret")

	t.Setenv("FUNCTION_CACHE_SIZE", "1")
	t.Setenv("FUNCTION_CACHE_TTL", "0")
	tmpDir := t.TempDir()
	t.Setenv("FUNCTION_CACHE_BIN_DIR", tmpDir)
	config.LoadConfig()

	cache, err := NewMetadataBinaryCache()
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
	t.Setenv("JWT_EXPIRATION_TIME", "3600")
	t.Setenv("JWT_PEM_PATH", "../.devcontainer/private.pem")
	t.Setenv("MIDDLEWARE_SECRET", "very VERY secret")
	t.Setenv("ERMES_SECRET", "very secret")
	t.Setenv("DEV_SECRET", "super secret")

	t.Setenv("FUNCTION_CACHE_SIZE", "20")
	t.Setenv("FUNCTION_CACHE_TTL", strconv.FormatInt((1*time.Second).Nanoseconds(), 10))
	tmpDir := t.TempDir()
	t.Setenv("FUNCTION_CACHE_BIN_DIR", tmpDir)
	config.LoadConfig()

	cache, err := NewMetadataBinaryCache()
	require.NoError(t, err)

	// Nothing set yet
	m, ok := cache.Get("missing", "missing")
	require.NoError(t, err)
	require.False(t, ok)
	require.Zero(t, m)
}
