package storage

import (
	"sync"
	"time"
)

type StorageStatistics struct {
	BlocksSaved       uint64
	TransactionsSaved uint64
	UTXOEntries       uint64
	CacheHits         uint64
	CacheMisses       uint64
	BytesWritten      uint64
	BytesRead         uint64
	SnapshotsCreated  uint64
	Flushes           uint64
	Compactions       uint64
	AverageLatency    time.Duration

	mu sync.RWMutex
}

func (s *StorageStatistics) Snapshot() StorageStatistics {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return StorageStatistics{
		BlocksSaved:       s.BlocksSaved,
		TransactionsSaved: s.TransactionsSaved,
		UTXOEntries:       s.UTXOEntries,
		CacheHits:         s.CacheHits,
		CacheMisses:       s.CacheMisses,
		BytesWritten:      s.BytesWritten,
		BytesRead:         s.BytesRead,
		SnapshotsCreated:  s.SnapshotsCreated,
		Flushes:           s.Flushes,
		Compactions:       s.Compactions,
		AverageLatency:    s.AverageLatency,
	}
}

func (s *StorageStatistics) IncBlocksSaved() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.BlocksSaved++
}

func (s *StorageStatistics) IncTransactionsSaved() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.TransactionsSaved++
}

func (s *StorageStatistics) UpdateUTXOEntries(count uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.UTXOEntries = count
}

func (s *StorageStatistics) IncCacheHits() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.CacheHits++
}

func (s *StorageStatistics) IncCacheMisses() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.CacheMisses++
}

func (s *StorageStatistics) AddBytesWritten(n uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.BytesWritten += n
}

func (s *StorageStatistics) AddBytesRead(n uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.BytesRead += n
}

func (s *StorageStatistics) IncSnapshotsCreated() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.SnapshotsCreated++
}

func (s *StorageStatistics) IncFlushes() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Flushes++
}

func (s *StorageStatistics) IncCompactions() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Compactions++
}

func (s *StorageStatistics) RecordLatency(d time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.AverageLatency == 0 {
		s.AverageLatency = d
	} else {
		s.AverageLatency = (s.AverageLatency + d) / 2
	}
}
