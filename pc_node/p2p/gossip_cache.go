package p2p

import (
	"context"
	"sync"
	"time"
)

type CacheEntry struct {
	Timestamp time.Time
	ExpiresAt time.Time
}

type GossipCache struct {
	entries    map[string]CacheEntry
	expiration time.Duration
	mu         sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
}

func NewGossipCache(expiration time.Duration) *GossipCache {
	ctx, cancel := context.WithCancel(context.Background())
	cache := &GossipCache{
		entries:    make(map[string]CacheEntry),
		expiration: expiration,
		ctx:        ctx,
		cancel:     cancel,
	}
	go cache.cleanupLoop()
	return cache
}

// Add adds a message ID to the cache. Returns true if added, false if already exists.
func (c *GossipCache) Add(msgID string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.entries[msgID]; exists {
		return false
	}

	c.entries[msgID] = CacheEntry{
		Timestamp: time.Now(),
		ExpiresAt: time.Now().Add(c.expiration),
	}
	return true
}

func (c *GossipCache) Contains(msgID string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	_, exists := c.entries[msgID]
	return exists
}

func (c *GossipCache) cleanupLoop() {
	ticker := time.NewTicker(c.expiration / 2)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			c.cleanup()
		}
	}
}

func (c *GossipCache) cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for id, entry := range c.entries {
		if now.After(entry.ExpiresAt) {
			delete(c.entries, id)
		}
	}
}

func (c *GossipCache) Stop() {
	c.cancel()
}

func (c *GossipCache) Count() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.entries)
}
