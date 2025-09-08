package caches

import (
	"caches/metadata"
	"ermes-central-repo/config"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/sync/singleflight"
)

// MetadataBinaryCache stores compiled binaries on disk and keeps track of them
// using an underlying BaseCache. Evicted binaries are automatically removed
// from disk.
type MetadataBinaryCache struct {
	cache   *BaseCache[metadata.FunctionMetadata]
	binDir  string
	fetcher Fetcher[*FunctionBundle]
	group   singleflight.Group
}

// NewMetadataBinaryCache creates a new MetadataBinaryCache based on the given configuration.
// Evicted binaries are automatically deleted from disk.
//
// Parameters:
//   - config: config.MetadataBinaryCacheConfig containing the cache size, TTL, and binary directory
//
// Returns:
//   - *MetadataBinaryCache: the newly created binary cache
//   - error: non-nil if the underlying BaseCache could not be created
func NewMetadataBinaryCache() (*MetadataBinaryCache, error) {
	binCache, err := NewBaseCacheWithEviction(config.FunctionCacheSize(), config.FunctionCacheTtl(), func(key string, item CacheItem[metadata.FunctionMetadata]) {
		_ = os.Remove(filepath.Join(config.FunctionCacheBinDir(), item.Value.Name+"@"+item.Value.Version))
	})
	if err != nil {
		return nil, err
	}

	return &MetadataBinaryCache{
		cache:   binCache,
		binDir:  config.FunctionCacheBinDir(),
		group:   singleflight.Group{},
		fetcher: NewRemoteBundleFetcher(fmt.Sprintf("http://%s:%s", config.FunctionRegistryAddress(), config.FunctionRegistryPort())),
	}, nil
}

// Set stores a binary on disk and records its path in the cache.
//
// Parameters:
//   - functionName: the name of the function associated with the binary
//   - binary: the binary content to store
//
// Returns:
//   - error: non-nil if writing the file to disk fails
func (c *MetadataBinaryCache) set(functionMetadata metadata.FunctionMetadata, binary []byte) error {
	functionNameString := functionMetadata.Name + "@" + functionMetadata.Version
	binPath := filepath.Join(c.binDir, functionNameString+".wasm")

	err := os.WriteFile(binPath, binary, 0644)
	if err != nil {
		return err
	}

	c.cache.Set(functionNameString, functionMetadata)
	return nil
}

func (c *MetadataBinaryCache) Get(functionName string, functionVersion string) (*metadata.FunctionMetadata, error) {
	functionNameString := functionName + "@" + functionVersion
	if fnMetadata, ok := c.cache.Get(functionNameString); ok {
		return &fnMetadata, nil
	} else {
		if fnBundle, err := c.fetchAndStore(functionName, functionVersion); err == nil {
			return &fnBundle.Metadata, nil
		} else {
			return nil, err
		}
	}
}

func (c *MetadataBinaryCache) fetchAndStore(functionName string, functionVersion string) (*FunctionBundle, error) {
	functionNameString := functionName + "@" + functionVersion
	fetchedValue, err, _ := c.group.Do(functionNameString, func() (any, error) {
		bundle, err := c.fetcher.Fetch(functionNameString)
		if err != nil {
			return nil, err
		}
		if bundle == nil || len(bundle.Binary) == 0 {
			return nil, fmt.Errorf("invalid bundle from registry")
		}

		if err = c.set(bundle.Metadata, bundle.Binary); err != nil {
			return nil, err
		}

		return bundle, nil
	})
	if err != nil {
		return nil, err
	}
	return fetchedValue.(*FunctionBundle), nil
}
