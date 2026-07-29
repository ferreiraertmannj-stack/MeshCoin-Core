package mempool

import (
	"sync"
	"time"
)

type dedupEntry struct {
	addedAt time.Time
}

type TransactionDedup struct {
	cache map[string]dedupEntry
	ttl   time.Duration
	mu    sync.RWMutex
}

func NewTransactionDedup(ttl time.Duration) *TransactionDedup {
	return &TransactionDedup{
		cache: make(map[string]dedupEntry),
		ttl:   ttl,
	}
}

// CheckAndAdd returns true if it's a NEW hash (and adds it), false if already exists
func (d *TransactionDedup) CheckAndAdd(hash string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	if _, exists := d.cache[hash]; exists {
		return false
	}

	d.cache[hash] = dedupEntry{addedAt: time.Now()}
	return true
}

// Contains returns true if the hash was already seen
func (d *TransactionDedup) Contains(hash string) bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	_, exists := d.cache[hash]
	return exists
}

// Cleanup removes entries older than the TTL
func (d *TransactionDedup) Cleanup() {
	d.mu.Lock()
	defer d.mu.Unlock()

	now := time.Now()
	for k, v := range d.cache {
		if now.Sub(v.addedAt) > d.ttl {
			delete(d.cache, k)
		}
	}
}
