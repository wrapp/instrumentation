package loader

import (
	"context"
	"time"

	dataloader "github.com/graph-gophers/dataloader"
	cache "github.com/patrickmn/go-cache"
)

// TTLCache implements the dataloader.Cache interface and is a thread-safe in-memory cache with expiration times
type TTLCache struct {
	c *cache.Cache
}

// NewTTLCache creates new TTLCache
func NewTTLCache(defaultExpiration, cleanupInterval time.Duration) dataloader.Cache {
	c := cache.New(defaultExpiration, cleanupInterval)
	cache := &TTLCache{c}
	return cache
}

// Get gets a value from the cache
func (c *TTLCache) Get(_ context.Context, key dataloader.Key) (dataloader.Thunk, bool) {
	v, ok := c.c.Get(key.String())
	if ok {
		return v.(dataloader.Thunk), ok
	}
	return nil, ok
}

// Set sets a value in the cache
func (c *TTLCache) Set(_ context.Context, key dataloader.Key, value dataloader.Thunk) {
	c.c.Set(key.String(), value, 0)
}

// Delete deletes and item in the cache
func (c *TTLCache) Delete(_ context.Context, key dataloader.Key) bool {
	if _, found := c.c.Get(key.String()); found {
		c.c.Delete(key.String())
		return true
	}
	return false
}

// Clear clears the cache
func (c *TTLCache) Clear() {
	c.c.Flush()
}
