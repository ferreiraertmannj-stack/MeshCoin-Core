package utxo

import (
	"sync"
	"time"
)

type cacheEntry struct {
	utxo      UTXO
	timestamp time.Time
}

type UTXOCache struct {
	cache map[OutPoint]cacheEntry
	ttl   time.Duration
	mu    sync.RWMutex
}

func NewUTXOCache(ttl time.Duration) *UTXOCache {
	return &UTXOCache{
		cache: make(map[OutPoint]cacheEntry),
		ttl:   ttl,
	}
}

func (c *UTXOCache) Get(op OutPoint) (UTXO, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, ok := c.cache[op]
	if !ok {
		return UTXO{}, false
	}
	if time.Since(entry.timestamp) > c.ttl {
		return UTXO{}, false
	}
	return entry.utxo, true
}

func (c *UTXOCache) Set(op OutPoint, utxo UTXO) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache[op] = cacheEntry{
		utxo:      utxo,
		timestamp: time.Now(),
	}
}

func (c *UTXOCache) Invalidate(op OutPoint) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.cache, op)
}

func (c *UTXOCache) Cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()
	now := time.Now()
	for k, v := range c.cache {
		if now.Sub(v.timestamp) > c.ttl {
			delete(c.cache, k)
		}
	}
}
