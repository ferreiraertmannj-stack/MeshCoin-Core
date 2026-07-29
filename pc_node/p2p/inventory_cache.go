package p2p

import (
	"context"
	"sync"
	"time"
)

type InventoryCacheEntry struct {
	Timestamp time.Time
	ExpiresAt time.Time
}

type InventoryCache struct {
	entries    map[string]InventoryCacheEntry
	expiration time.Duration
	mu         sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
}

func NewInventoryCache(expiration time.Duration) *InventoryCache {
	ctx, cancel := context.WithCancel(context.Background())
	cache := &InventoryCache{
		entries:    make(map[string]InventoryCacheEntry),
		expiration: expiration,
		ctx:        ctx,
		cancel:     cancel,
	}
	go cache.Cleanup()
	return cache
}

func (c *InventoryCache) Add(objectHash string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.entries[objectHash]; exists {
		return false
	}

	c.entries[objectHash] = InventoryCacheEntry{
		Timestamp: time.Now(),
		ExpiresAt: time.Now().Add(c.expiration),
	}
	return true
}

func (c *InventoryCache) Remove(objectHash string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.entries, objectHash)
}

func (c *InventoryCache) Contains(objectHash string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	_, exists := c.entries[objectHash]
	return exists
}

func (c *InventoryCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries = make(map[string]InventoryCacheEntry)
}

func (c *InventoryCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.entries)
}

func (c *InventoryCache) Cleanup() {
	ticker := time.NewTicker(c.expiration / 2)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			c.mu.Lock()
			now := time.Now()
			for hash, entry := range c.entries {
				if now.After(entry.ExpiresAt) {
					delete(c.entries, hash)
				}
			}
			c.mu.Unlock()
		}
	}
}

func (c *InventoryCache) Stop() {
	c.cancel()
}
