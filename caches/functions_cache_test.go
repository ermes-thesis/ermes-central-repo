package caches

import (
	"ermes-central-repo/config"
	common_caches "ermes-central-repo/mocks/ermes-common/caches"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ermes-thesis/ermes-common/caches"
	"github.com/ermes-thesis/ermes-common/logger"
	"github.com/ermes-thesis/ermes-common/model"
	"github.com/stretchr/testify/require"
)

func createNewLogger(t *testing.T) logger.Logger {
	log, err := logger.NewLogger(logger.Config{
		Level:     logger.DebugLevel,
		AddSource: false,
	})
	require.NoError(t, err)
	return log
}

func TestFunctionCacheImpl_GetBinary(t *testing.T) {
	log := createNewLogger(t)
	tmpDir := t.TempDir()

	fn := "authFunc1"
	fnVer := "1.2.3"

	fetcher := &common_caches.Fetcher[*FunctionBundle]{}

	fetcher.EXPECT().Fetch(fn, fnVer).Return(&FunctionBundle{
		Binary: []byte("binary-" + fn),
		Metadata: model.FunctionMetadata{
			Name:                  fn,
			Version:               fnVer,
			TemplateNames:         nil,
			DeployedOn:            nil,
			StrongCollWriteAccess: false,
		},
	}, nil)
	fetcher.EXPECT().Fetch(fn, "latest").Return(&FunctionBundle{
		Binary: []byte("binary-latest"),
		Metadata: model.FunctionMetadata{
			Name:                  fn,
			Version:               "1.2.10",
			TemplateNames:         nil,
			DeployedOn:            nil,
			StrongCollWriteAccess: false,
		},
	}, nil)

	cfg := &config.Config{
		FunctionCache: caches.MetadataBinaryCacheConfig{
			Size:   10,
			Ttl:    30 * time.Second,
			BinDir: tmpDir,
		},
		LatestCache: caches.LatestCacheConfig{
			Size: 10,
			Ttl:  30 * time.Second,
		},
	}

	cache, err := NewFunctionCache(cfg, fetcher, log)
	require.NoError(t, err)

	// Initially cache is empty, so fetcher should be called
	bin, err := cache.GetBinary(fn, fnVer)
	require.NoError(t, err)
	require.Equal(t, []byte("binary-"+fn), bin)
	// Assert that Fetch has been called 1 time
	fetcher.AssertNumberOfCalls(t, "Fetch", 1)

	// Requesting binary for the same fn. Fetcher SHOULD NOT BE CALLED
	bin, err = cache.GetBinary(fn, fnVer)
	require.NoError(t, err)
	require.Equal(t, []byte("binary-"+fn), bin)
	// Assert that Fetch still, has been called just 1 time
	fetcher.AssertNumberOfCalls(t, "Fetch", 1)

	// Check that binary file exists on disk
	binPath := filepath.Join(tmpDir, fn+"@"+fnVer+".wasm")
	if _, err := os.Stat(binPath); os.IsNotExist(err) {
		require.Fail(t, fmt.Sprintf("%s does not exist", binPath))
	}

	// Testing "latest" behaviour
	// Try to get same binary but latest version
	bin, err = cache.GetBinary(fn, "latest")
	require.NoError(t, err)
	require.Equal(t, []byte("binary-latest"), bin)
	// Since the latest version is different from the one in the cache, Fetch should be called
	fetcher.AssertNumberOfCalls(t, "Fetch", 2)

	// Try to get same binary but latest version again
	bin, err = cache.GetBinary(fn, "latest")
	require.NoError(t, err)
	require.Equal(t, []byte("binary-latest"), bin)
	// latest is in cache, Fetch should not be called
	fetcher.AssertNumberOfCalls(t, "Fetch", 2)

	// Check that binary file exists on disk
	binPath = filepath.Join(tmpDir, fn+"@1.2.10.wasm")
	if _, err := os.Stat(binPath); os.IsNotExist(err) {
		require.Fail(t, fmt.Sprintf("%s does not exist", binPath))
	}
}

func TestFunctionCacheImpl_GetMetadata(t *testing.T) {
	log := createNewLogger(t)
	tmpDir := t.TempDir()

	fn := "authFunc1"
	fnVer := "1.2.3"

	fetcher := &common_caches.Fetcher[*FunctionBundle]{}

	fetcher.EXPECT().Fetch(fn, fnVer).Return(&FunctionBundle{
		Binary: []byte("binary-" + fn),
		Metadata: model.FunctionMetadata{
			Name:                  fn,
			Version:               fnVer,
			TemplateNames:         nil,
			DeployedOn:            []string{"IT"},
			StrongCollWriteAccess: false,
			IsAuth:                true,
		},
	}, nil)
	fetcher.EXPECT().Fetch(fn, "latest").Return(&FunctionBundle{
		Binary: []byte("binary-latest"),
		Metadata: model.FunctionMetadata{
			Name:                  fn,
			Version:               "1.2.10",
			TemplateNames:         nil,
			DeployedOn:            []string{"FR", "IT"},
			StrongCollWriteAccess: true,
			IsAuth:                true,
		},
	}, nil)

	cfg := &config.Config{
		FunctionCache: caches.MetadataBinaryCacheConfig{
			Size:   10,
			Ttl:    30 * time.Second,
			BinDir: tmpDir,
		},
		LatestCache: caches.LatestCacheConfig{
			Size: 10,
			Ttl:  30 * time.Second,
		},
	}

	cache, err := NewFunctionCache(cfg, fetcher, log)
	require.NoError(t, err)

	// Initially cache is empty, so fetcher should be called
	meta, err := cache.GetMetadata(fn, fnVer)
	require.NoError(t, err)
	require.Equal(t, fn, meta.Name)
	require.Equal(t, fnVer, meta.Version)
	require.Equal(t, []string{"IT"}, meta.DeployedOn)
	require.False(t, meta.StrongCollWriteAccess)
	require.True(t, meta.IsAuth)
	// Assert that Fetch has been called 1 time
	fetcher.AssertNumberOfCalls(t, "Fetch", 1)

	// Requesting binary for the same fn. Fetcher SHOULD NOT BE CALLED
	meta2, err := cache.GetMetadata(fn, fnVer)
	require.NoError(t, err)
	require.Equal(t, meta2.Name, meta.Name)
	require.Equal(t, meta2.Version, meta.Version)
	require.Equal(t, meta2.DeployedOn, meta.DeployedOn)
	require.False(t, meta2.StrongCollWriteAccess)
	require.True(t, meta2.IsAuth)
	// Assert that Fetch still, has been called just 1 time
	fetcher.AssertNumberOfCalls(t, "Fetch", 1)

	// Check that binary file exists on disk
	binPath := filepath.Join(tmpDir, fn+"@"+fnVer+".wasm")
	if _, err := os.Stat(binPath); os.IsNotExist(err) {
		require.Fail(t, fmt.Sprintf("%s does not exist", binPath))
	}

	// Try to get same binary but latest version
	metaLast, err := cache.GetMetadata(fn, "latest")
	require.NoError(t, err)
	require.Equal(t, fn, metaLast.Name)
	require.Equal(t, "1.2.10", metaLast.Version)
	require.Equal(t, []string{"FR", "IT"}, metaLast.DeployedOn)
	require.True(t, metaLast.StrongCollWriteAccess)
	require.True(t, metaLast.IsAuth)
	// Since the latest version is different from the one in the cache, Fetch should be called
	fetcher.AssertNumberOfCalls(t, "Fetch", 2)

	// Try to get same binary but latest version again
	metaLast2, err := cache.GetMetadata(fn, "latest")
	require.NoError(t, err)
	require.Equal(t, metaLast2.Name, metaLast.Name)
	require.Equal(t, metaLast2.Version, metaLast.Version)
	require.Equal(t, metaLast2.DeployedOn, metaLast.DeployedOn)
	require.True(t, metaLast2.StrongCollWriteAccess)
	require.True(t, metaLast2.IsAuth)
	// latest is in cache, Fetch should not be called
	fetcher.AssertNumberOfCalls(t, "Fetch", 2)

	// Check that binary file exists on disk
	binPath = filepath.Join(tmpDir, fn+"@1.2.10.wasm")
	if _, err := os.Stat(binPath); os.IsNotExist(err) {
		require.Fail(t, fmt.Sprintf("%s does not exist", binPath))
	}
}

func TestFunctionCacheImpl_resolveLatest(t *testing.T) {
	log := createNewLogger(t)
	tmpDir := t.TempDir()

	fn := "authFunc1"

	fetcher := &common_caches.Fetcher[*FunctionBundle]{}

	fetcher.EXPECT().Fetch(fn, "latest").Return(&FunctionBundle{
		Binary: []byte("binary-latest"),
		Metadata: model.FunctionMetadata{
			Name:                  fn,
			Version:               "1.2.10",
			TemplateNames:         nil,
			DeployedOn:            []string{"FR", "IT"},
			StrongCollWriteAccess: true,
			IsAuth:                true,
		},
	}, nil)
	fetcher.EXPECT().FetchLatest(fn, "1.2.10").Return(&FunctionBundle{
		Binary: []byte("binary-more-latest"),
		Metadata: model.FunctionMetadata{
			Name:                  fn,
			Version:               "1.2.15",
			TemplateNames:         nil,
			DeployedOn:            []string{"FR", "IT", "KR"},
			StrongCollWriteAccess: true,
			IsAuth:                true,
		},
	}, nil)

	cfg := &config.Config{
		FunctionCache: caches.MetadataBinaryCacheConfig{
			Size:   10,
			Ttl:    0, // Disable ttl eviction for MetadataBinaryCache
			BinDir: tmpDir,
		},
		LatestCache: caches.LatestCacheConfig{
			Size: 10,
			Ttl:  1 * time.Second,
		},
	}

	cache, err := NewFunctionCache(cfg, fetcher, log)
	require.NoError(t, err)

	// This should call the Fetch method, since the cache is empty
	fnBundle, err := cache.resolveLatest(fn, "binary")
	require.NoError(t, err)
	require.Equal(t, []byte("binary-latest"), fnBundle.Binary)
	fetcher.AssertNumberOfCalls(t, "Fetch", 1)
	fetcher.AssertNumberOfCalls(t, "FetchLatest", 0)

	time.Sleep(2 * time.Second)

	// The latest cached should be expired. FetchLatest should be called
	fnBundle, err = cache.resolveLatest(fn, "binary")
	require.NoError(t, err)
	require.Equal(t, []byte("binary-more-latest"), fnBundle.Binary)
	fetcher.AssertNumberOfCalls(t, "Fetch", 1)
	fetcher.AssertNumberOfCalls(t, "FetchLatest", 1)
}

func TestFunctionCacheImpl_GetBinaryFail(t *testing.T) {
	log := createNewLogger(t)
	tmpDir := t.TempDir()

	fn := "authFunc1"

	fetcher := &common_caches.Fetcher[*FunctionBundle]{}

	fetcher.EXPECT().Fetch(fn, "latest").Return(nil, errors.New("some error"))
	fetcher.EXPECT().Fetch(fn, "1.0.0").Return(&FunctionBundle{
		Binary: []byte("valid binary"),
		Metadata: model.FunctionMetadata{
			Name:                  fn,
			Version:               "1.0.0",
			TemplateNames:         nil,
			DeployedOn:            nil,
			StrongCollWriteAccess: true,
			IsAuth:                true,
		},
	}, nil)
	fetcher.EXPECT().Fetch(fn, "10.1.1").Return(nil, errors.New("some error"))

	cfg := &config.Config{
		FunctionCache: caches.MetadataBinaryCacheConfig{
			Size:   10,
			Ttl:    0, // Disable ttl eviction for MetadataBinaryCache
			BinDir: tmpDir,
		},
		LatestCache: caches.LatestCacheConfig{
			Size: 10,
			Ttl:  1 * time.Second,
		},
	}

	cache, err := NewFunctionCache(cfg, fetcher, log)
	require.NoError(t, err)

	_, err = cache.GetBinary(fn, "latest")
	require.EqualError(t, err, "some error")

	// Should give error if reading of the file fails
	_, err = cache.GetBinary(fn, "1.0.0")
	require.NoError(t, err)
	err = os.Remove(filepath.Join(tmpDir, fn+"@1.0.0.wasm"))
	_, err = cache.GetBinary(fn, "1.0.0")
	require.Error(t, err)

	// FetchAndStore gives error
	_, err = cache.GetBinary(fn, "10.1.1")
	require.EqualError(t, err, "some error")
}

func TestFunctionCacheImpl_GetMetadataFail(t *testing.T) {
	log := createNewLogger(t)
	tmpDir := t.TempDir()

	fn := "authFunc1"

	fetcher := &common_caches.Fetcher[*FunctionBundle]{}

	fetcher.EXPECT().Fetch(fn, "latest").Return(nil, errors.New("some error"))
	fetcher.EXPECT().Fetch(fn, "1.0.0").Return(&FunctionBundle{
		Binary: []byte("valid binary"),
		Metadata: model.FunctionMetadata{
			Name:                  fn,
			Version:               "1.0.0",
			TemplateNames:         nil,
			DeployedOn:            nil,
			StrongCollWriteAccess: true,
			IsAuth:                true,
		},
	}, nil)
	fetcher.EXPECT().Fetch(fn, "10.1.1").Return(nil, errors.New("some error"))

	cfg := &config.Config{
		FunctionCache: caches.MetadataBinaryCacheConfig{
			Size:   10,
			Ttl:    0, // Disable ttl eviction for MetadataBinaryCache
			BinDir: tmpDir,
		},
		LatestCache: caches.LatestCacheConfig{
			Size: 10,
			Ttl:  1 * time.Second,
		},
	}

	cache, err := NewFunctionCache(cfg, fetcher, log)
	require.NoError(t, err)

	_, err = cache.GetMetadata(fn, "latest")
	require.EqualError(t, err, "some error")

	// FetchAndStore gives error
	_, err = cache.GetMetadata(fn, "10.1.1")
	require.EqualError(t, err, "some error")
}

func TestFetchAndStore_InvalidBundle(t *testing.T) {
	log := createNewLogger(t)
	tmpDir := t.TempDir()

	fn := "authFunc1"

	fetcher := &common_caches.Fetcher[*FunctionBundle]{}

	fetcher.EXPECT().Fetch(fn, "1.0.0").Return(&FunctionBundle{
		Binary: []byte{},
		Metadata: model.FunctionMetadata{
			Name:                  fn,
			Version:               "1.0.0",
			TemplateNames:         nil,
			DeployedOn:            nil,
			StrongCollWriteAccess: true,
			IsAuth:                true,
		},
	}, nil)

	cfg := &config.Config{
		FunctionCache: caches.MetadataBinaryCacheConfig{
			Size:   10,
			Ttl:    0, // Disable ttl eviction for MetadataBinaryCache
			BinDir: tmpDir,
		},
		LatestCache: caches.LatestCacheConfig{
			Size: 10,
			Ttl:  1 * time.Second,
		},
	}

	cache, err := NewFunctionCache(cfg, fetcher, log)
	require.NoError(t, err)

	_, err = cache.fetchAndStore(fn, "1.0.0")
	require.EqualError(t, err, "invalid bundle from registry")
}
