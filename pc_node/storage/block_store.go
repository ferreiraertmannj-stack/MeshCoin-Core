package storage

import (
	"fmt"
	"sync"
)

type Block struct {
	Hash   string
	Height uint64
	Parent string
	Data   []byte
}

type BlockStore struct {
	mu       sync.RWMutex
	byHash   map[string]*Block
	byHeight map[uint64]*Block
	stats    *StorageStatistics
	cache    *StorageCache
}

func NewBlockStore(stats *StorageStatistics, cache *StorageCache) *BlockStore {
	return &BlockStore{
		byHash:   make(map[string]*Block),
		byHeight: make(map[uint64]*Block),
		stats:    stats,
		cache:    cache,
	}
}

func (s *BlockStore) Save(b *Block) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.byHash[b.Hash] = b
	s.byHeight[b.Height] = b

	s.stats.IncBlocksSaved()
	s.stats.AddBytesWritten(uint64(len(b.Data)))

	// Pre-warm cache
	s.cache.Set("block_"+b.Hash, b)
	s.cache.Set(fmt.Sprintf("block_%d", b.Height), b)

	return nil
}

func (s *BlockStore) LoadByHash(hash string) (*Block, error) {
	key := "block_" + hash
	if cached, ok := s.cache.Get(key); ok {
		s.stats.IncCacheHits()
		return cached.(*Block), nil
	}
	s.stats.IncCacheMisses()

	s.mu.RLock()
	defer s.mu.RUnlock()
	b, ok := s.byHash[hash]
	if !ok {
		return nil, fmt.Errorf("block not found by hash: %s", hash)
	}
	s.cache.Set(key, b)
	s.stats.AddBytesRead(uint64(len(b.Data)))
	return b, nil
}

func (s *BlockStore) LoadByHeight(height uint64) (*Block, error) {
	key := fmt.Sprintf("block_%d", height)
	if cached, ok := s.cache.Get(key); ok {
		s.stats.IncCacheHits()
		return cached.(*Block), nil
	}
	s.stats.IncCacheMisses()

	s.mu.RLock()
	defer s.mu.RUnlock()
	b, ok := s.byHeight[height]
	if !ok {
		return nil, fmt.Errorf("block not found by height: %d", height)
	}
	s.cache.Set(key, b)
	s.stats.AddBytesRead(uint64(len(b.Data)))
	return b, nil
}
