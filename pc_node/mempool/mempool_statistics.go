package mempool

import (
	"sync"
	"time"
)

type MempoolStatistics struct {
	Transactions          uint64
	Bytes                 uint64
	Hits                  uint64
	Misses                uint64
	Duplicates            uint64
	Expired               uint64
	Dropped               uint64
	AverageLifetime       time.Duration
	AverageFee            uint64
	AverageValidationTime time.Duration

	mu sync.RWMutex
}

func (s *MempoolStatistics) Snapshot() MempoolStatistics {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return MempoolStatistics{
		Transactions:          s.Transactions,
		Bytes:                 s.Bytes,
		Hits:                  s.Hits,
		Misses:                s.Misses,
		Duplicates:            s.Duplicates,
		Expired:               s.Expired,
		Dropped:               s.Dropped,
		AverageLifetime:       s.AverageLifetime,
		AverageFee:            s.AverageFee,
		AverageValidationTime: s.AverageValidationTime,
	}
}

func (s *MempoolStatistics) IncTransactions(bytes uint64, fee uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Moving average calculation for fee (simplified)
	if s.Transactions == 0 {
		s.AverageFee = fee
	} else {
		s.AverageFee = (s.AverageFee*s.Transactions + fee) / (s.Transactions + 1)
	}

	s.Transactions++
	s.Bytes += bytes
}

func (s *MempoolStatistics) DecTransactions(bytes uint64, lifetime time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.Transactions > 0 {
		s.Transactions--
	}
	if s.Bytes >= bytes {
		s.Bytes -= bytes
	} else {
		s.Bytes = 0
	}

	// Update average lifetime
	if s.AverageLifetime == 0 {
		s.AverageLifetime = lifetime
	} else {
		// Simple moving average
		s.AverageLifetime = (s.AverageLifetime + lifetime) / 2
	}
}

func (s *MempoolStatistics) IncHits() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Hits++
}

func (s *MempoolStatistics) IncMisses() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Misses++
}

func (s *MempoolStatistics) IncDuplicates() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Duplicates++
}

func (s *MempoolStatistics) IncExpired() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Expired++
}

func (s *MempoolStatistics) IncDropped() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Dropped++
}

func (s *MempoolStatistics) RecordValidationTime(d time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.AverageValidationTime == 0 {
		s.AverageValidationTime = d
	} else {
		s.AverageValidationTime = (s.AverageValidationTime + d) / 2
	}
}
