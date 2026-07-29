package mining

import (
	"sync"
	"time"
)

type MiningCache struct {
	job      *MiningJob
	cachedAt time.Time
	ttl      time.Duration
	mu       sync.RWMutex
}

func NewMiningCache(ttl time.Duration) *MiningCache {
	return &MiningCache{
		ttl: ttl,
	}
}

func (c *MiningCache) Set(job *MiningJob) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.job = job
	c.cachedAt = time.Now()
}

func (c *MiningCache) Get() (*MiningJob, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.job == nil {
		return nil, false
	}

	if time.Since(c.cachedAt) > c.ttl {
		return nil, false // expired
	}

	return c.job, true
}

func (c *MiningCache) Invalidate() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.job = nil
}
