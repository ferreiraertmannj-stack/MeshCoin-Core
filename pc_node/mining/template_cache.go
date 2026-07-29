package mining

import (
	"sync"
	"time"
)

type TemplateCache struct {
	template *BlockTemplate
	cachedAt time.Time
	ttl      time.Duration
	mu       sync.RWMutex
}

func NewTemplateCache(ttl time.Duration) *TemplateCache {
	return &TemplateCache{
		ttl: ttl,
	}
}

func (c *TemplateCache) Set(tmpl *BlockTemplate) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.template = tmpl
	c.cachedAt = time.Now()
}

func (c *TemplateCache) Get() (*BlockTemplate, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.template == nil {
		return nil, false
	}

	if time.Since(c.cachedAt) > c.ttl {
		return nil, false // expired
	}

	return c.template, true
}

func (c *TemplateCache) Invalidate() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.template = nil
}
