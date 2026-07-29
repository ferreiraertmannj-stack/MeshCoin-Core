package storage

import (
	"container/list"
	"sync"
	"time"
)

type cacheEntry struct {
	key       string
	value     interface{}
	timestamp time.Time
}

type StorageCache struct {
	capacity  int
	ttl       time.Duration
	mu        sync.RWMutex
	items     map[string]*list.Element
	evictList *list.List
}

func NewStorageCache(capacity int, ttl time.Duration) *StorageCache {
	return &StorageCache{
		capacity:  capacity,
		ttl:       ttl,
		items:     make(map[string]*list.Element),
		evictList: list.New(),
	}
}

func (c *StorageCache) Get(key string) (interface{}, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	ent, ok := c.items[key]
	if !ok {
		return nil, false
	}

	entry := ent.Value.(*cacheEntry)
	if time.Since(entry.timestamp) > c.ttl {
		c.evictList.Remove(ent)
		delete(c.items, key)
		return nil, false
	}

	c.evictList.MoveToFront(ent)
	return entry.value, true
}

func (c *StorageCache) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if ent, ok := c.items[key]; ok {
		c.evictList.MoveToFront(ent)
		entry := ent.Value.(*cacheEntry)
		entry.value = value
		entry.timestamp = time.Now()
		return
	}

	entry := &cacheEntry{key: key, value: value, timestamp: time.Now()}
	ent := c.evictList.PushFront(entry)
	c.items[key] = ent

	if c.capacity != 0 && c.evictList.Len() > c.capacity {
		c.removeOldest()
	}
}

func (c *StorageCache) removeOldest() {
	ent := c.evictList.Back()
	if ent != nil {
		c.evictList.Remove(ent)
		entry := ent.Value.(*cacheEntry)
		delete(c.items, entry.key)
	}
}

func (c *StorageCache) Remove(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if ent, ok := c.items[key]; ok {
		c.evictList.Remove(ent)
		delete(c.items, key)
	}
}
