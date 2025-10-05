package caches

import (
	"sync"
	"time"

	"github.com/hashicorp/golang-lru/v2"
)

// GenericCache defines the behaviour for a key-value cache that supports storing
// and retrieving values, with optional TTL-based expiration.
//
// Type Parameters:
//   - T: the type of the cached values.
type GenericCache[T any] interface {
	// Get retrieves the value associated with the given key.
	//
	// Parameters:
	//   - key: the key to look up in the cache
	//
	// Returns:
	//   - T: the stored value, or the zero value if the key is missing or expired.
	//   - bool: true if the key exists and the value is not expired, false otherwise
	Get(key string) (T, bool)

	// GetWithStatus retrieves the value for a key along with status flags indicating
	// whether the key exists and whether it is expired.
	//
	// Parameters:
	//	 - key: the key to look up in the cache.
	//
	// Returns:
	//	 - T: the stored value, or the zero value if the key is missing.
	//	 - bool: true if the key exists in the cache (even if expired), false otherwise
	//	 - bool: true if the key exists but the value is expired, false otherwise.
	GetWithStatus(key string) (T, bool, bool)

	// Set stores the given value under the specified key.
	//
	// Parameters:
	//   - key: the key under which to store the value
	//   - value: the value to store
	Set(key string, value T)
}

// CacheItem represents a cached value along with its insertion timestamp.
// The timestamp is used to enforce TTL (time-to-live) expiration policies.
type CacheItem[T any] struct {
	Value     T
	Timestamp time.Time
}

// BaseCache is a GenericCache implementation with LRU eviction and optional TTL.
//
// If ttl is zero, items never expire and eviction is determined solely
// by the LRU policy.
// It is safe for concurrent use by multiple goroutines
type BaseCache[T any] struct {
	cache *lru.Cache[string, CacheItem[T]]
	ttl   time.Duration
	mu    sync.RWMutex
}

// NewBaseCache creates a new BaseCache with the given maximum size and TTL.
//
// Parameters:
//   - size: the maximum number of entries before LRU eviction
//   - ttl: the time-to-live for items (zero disables expiration)
//
// Returns:
//   - *BaseCache[T]: the newly created cache
//   - error: non-nil if the cache could not be created
func NewBaseCache[T any](size int, ttl time.Duration) (*BaseCache[T], error) {
	cache, err := lru.New[string, CacheItem[T]](size)
	if err != nil {
		return nil, err
	}
	return &BaseCache[T]{
		cache: cache,
		ttl:   ttl,
	}, nil
}

// NewBaseCacheWithEviction creates a BaseCache with the given size, TTL, and
// a custom eviction callback.
//
// Parameters:
//   - size: the maximum number of entries before LRU eviction
//   - ttl: the time-to-live for items (zero disables expiration)
//   - evictFn: function called whenever an item is evicted (either due to LRU policy or manually removed)
//
// Returns:
//   - *BaseCache[T]: the newly created cache
//   - error: non-nil if the cache could not be created
func NewBaseCacheWithEviction[T any](size int, ttl time.Duration, evictFn func(string, CacheItem[T])) (*BaseCache[T], error) {
	cache, err := lru.NewWithEvict[string, CacheItem[T]](size, evictFn)
	if err != nil {
		return nil, err
	}

	return &BaseCache[T]{
		cache: cache,
		ttl:   ttl,
	}, nil
}

// Get retrieves a value if present and not expired. Expired items are removed.
func (c *BaseCache[T]) Get(key string) (T, bool) {
	var zero T
	// Lock the cache in read
	c.mu.RLock()

	// Get the CacheItem
	if item, ok := c.cache.Get(key); ok {
		// If TTL is disabled or the item is not expired return it
		// and unlock the cache
		if c.ttl == 0 || time.Since(item.Timestamp) < c.ttl {
			c.mu.RUnlock()
			return item.Value, true
		}

		// If item is expired, upgrade the lock to a write lock
		c.mu.RUnlock()
		// Here, other threads can interact with the cache
		c.mu.Lock()
		defer c.mu.Unlock()

		// Perform again the same check to avoid concurrency problems
		if item, ok = c.cache.Get(key); ok {
			// If situation is as before, remove the item
			if time.Since(item.Timestamp) > c.ttl {
				c.cache.Remove(key)
				return zero, false
			} else {
				// Item was refreshed
				return item.Value, true
			}
		}
		return zero, false
	}

	// Unlock the cache from read
	c.mu.RUnlock()
	return zero, false
}

// Set stores a value under a key, with its insertion time.
func (c *BaseCache[T]) Set(key string, value T) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache.Add(key, CacheItem[T]{
		Value:     value,
		Timestamp: time.Now(),
	})
}

// GetWithStatus retrieves a value along with its expiration status.
//
// Unlike Get, this method does not remove expired items from the cache;
// it simply reports their status.
func (c *BaseCache[T]) GetWithStatus(key string) (T, bool, bool) {
	var zero T

	// Lock the cache in read
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Check if the item is present
	item, ok := c.cache.Get(key)
	if !ok {
		return zero, false, false
	}

	// If items is present, return it and whether is expired or not
	if c.ttl == 0 || time.Since(item.Timestamp) < c.ttl {
		return item.Value, true, false
	}
	return item.Value, true, true
}
