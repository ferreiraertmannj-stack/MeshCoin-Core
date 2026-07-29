package storage

import (
	"fmt"
	"sync"
)

type UTXOEntry struct {
	Outpoint string
	Data     []byte
}

type UTXOStore struct {
	mu    sync.RWMutex
	utxos map[string]*UTXOEntry
	stats *StorageStatistics
	cache *StorageCache
}

func NewUTXOStore(stats *StorageStatistics, cache *StorageCache) *UTXOStore {
	return &UTXOStore{
		utxos: make(map[string]*UTXOEntry),
		stats: stats,
		cache: cache,
	}
}

func (s *UTXOStore) Save(u *UTXOEntry) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.utxos[u.Outpoint] = u

	s.stats.UpdateUTXOEntries(uint64(len(s.utxos)))
	s.stats.AddBytesWritten(uint64(len(u.Data)))

	s.cache.Set("utxo_"+u.Outpoint, u)
	return nil
}

func (s *UTXOStore) Delete(outpoint string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.utxos, outpoint)
	s.stats.UpdateUTXOEntries(uint64(len(s.utxos)))
	s.cache.Remove("utxo_" + outpoint)
	return nil
}

func (s *UTXOStore) Load(outpoint string) (*UTXOEntry, error) {
	key := "utxo_" + outpoint
	if cached, ok := s.cache.Get(key); ok {
		s.stats.IncCacheHits()
		return cached.(*UTXOEntry), nil
	}
	s.stats.IncCacheMisses()

	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.utxos[outpoint]
	if !ok {
		return nil, fmt.Errorf("UTXO not found: %s", outpoint)
	}

	s.cache.Set(key, u)
	s.stats.AddBytesRead(uint64(len(u.Data)))
	return u, nil
}
