package storage

import (
	"fmt"
	"sync"
)

type Transaction struct {
	Hash      string
	BlockHash string
	Sender    string
	Receiver  string
	Data      []byte
}

type TransactionStore struct {
	mu     sync.RWMutex
	byHash map[string]*Transaction
	stats  *StorageStatistics
	cache  *StorageCache
}

func NewTransactionStore(stats *StorageStatistics, cache *StorageCache) *TransactionStore {
	return &TransactionStore{
		byHash: make(map[string]*Transaction),
		stats:  stats,
		cache:  cache,
	}
}

func (s *TransactionStore) Save(tx *Transaction) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.byHash[tx.Hash] = tx

	s.stats.IncTransactionsSaved()
	s.stats.AddBytesWritten(uint64(len(tx.Data)))

	s.cache.Set("tx_"+tx.Hash, tx)
	return nil
}

func (s *TransactionStore) Load(hash string) (*Transaction, error) {
	key := "tx_" + hash
	if cached, ok := s.cache.Get(key); ok {
		s.stats.IncCacheHits()
		return cached.(*Transaction), nil
	}
	s.stats.IncCacheMisses()

	s.mu.RLock()
	defer s.mu.RUnlock()
	tx, ok := s.byHash[hash]
	if !ok {
		return nil, fmt.Errorf("transaction not found: %s", hash)
	}
	s.cache.Set(key, tx)
	s.stats.AddBytesRead(uint64(len(tx.Data)))
	return tx, nil
}
