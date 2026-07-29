package utxo

import (
	"sync"
	"time"
)

type UTXOStatistics struct {
	UTXOsCreated          uint64
	UTXOsSpent            uint64
	TransactionsValidated uint64
	TransactionsRejected  uint64
	DoubleSpends          uint64
	CacheHits             uint64
	CacheMisses           uint64
	AverageValidationTime time.Duration
	SnapshotsCreated      uint64
	RollbacksExecuted     uint64

	mu sync.RWMutex
}

func (s *UTXOStatistics) Snapshot() UTXOStatistics {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return UTXOStatistics{
		UTXOsCreated:          s.UTXOsCreated,
		UTXOsSpent:            s.UTXOsSpent,
		TransactionsValidated: s.TransactionsValidated,
		TransactionsRejected:  s.TransactionsRejected,
		DoubleSpends:          s.DoubleSpends,
		CacheHits:             s.CacheHits,
		CacheMisses:           s.CacheMisses,
		AverageValidationTime: s.AverageValidationTime,
		SnapshotsCreated:      s.SnapshotsCreated,
		RollbacksExecuted:     s.RollbacksExecuted,
	}
}

func (s *UTXOStatistics) IncUTXOCreated(n uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.UTXOsCreated += n
}

func (s *UTXOStatistics) IncUTXOSpent(n uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.UTXOsSpent += n
}

func (s *UTXOStatistics) IncTransactionsValidated() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.TransactionsValidated++
}

func (s *UTXOStatistics) IncTransactionsRejected() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.TransactionsRejected++
}

func (s *UTXOStatistics) IncDoubleSpends() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.DoubleSpends++
}

func (s *UTXOStatistics) IncCacheHits() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.CacheHits++
}

func (s *UTXOStatistics) IncCacheMisses() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.CacheMisses++
}

func (s *UTXOStatistics) RecordValidationTime(d time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.AverageValidationTime == 0 {
		s.AverageValidationTime = d
	} else {
		s.AverageValidationTime = (s.AverageValidationTime + d) / 2
	}
}

func (s *UTXOStatistics) IncSnapshotsCreated() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.SnapshotsCreated++
}

func (s *UTXOStatistics) IncRollbacksExecuted() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.RollbacksExecuted++
}
