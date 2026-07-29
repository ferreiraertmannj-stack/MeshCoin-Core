package mining

import (
	"sync"
	"time"
)

type MiningStatistics struct {
	HashesComputed     uint64
	HashesPerSecond    float64
	SharesFound        uint64
	BlocksFound        uint64
	JobsCreated        uint64
	JobsCancelled      uint64
	AverageHashTime    time.Duration
	AverageJobLifetime time.Duration
	WorkerUtilization  float64
	CacheHits          uint64
	CacheMisses        uint64

	mu sync.RWMutex
}

func (s *MiningStatistics) Snapshot() MiningStatistics {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return MiningStatistics{
		HashesComputed:     s.HashesComputed,
		HashesPerSecond:    s.HashesPerSecond,
		SharesFound:        s.SharesFound,
		BlocksFound:        s.BlocksFound,
		JobsCreated:        s.JobsCreated,
		JobsCancelled:      s.JobsCancelled,
		AverageHashTime:    s.AverageHashTime,
		AverageJobLifetime: s.AverageJobLifetime,
		WorkerUtilization:  s.WorkerUtilization,
		CacheHits:          s.CacheHits,
		CacheMisses:        s.CacheMisses,
	}
}

func (s *MiningStatistics) AddHashes(n uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.HashesComputed += n
}

func (s *MiningStatistics) IncSharesFound() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.SharesFound++
}

func (s *MiningStatistics) IncBlocksFound() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.BlocksFound++
}

func (s *MiningStatistics) IncJobsCreated() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.JobsCreated++
}

func (s *MiningStatistics) IncJobsCancelled() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.JobsCancelled++
}

func (s *MiningStatistics) IncCacheHits() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.CacheHits++
}

func (s *MiningStatistics) IncCacheMisses() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.CacheMisses++
}

func (s *MiningStatistics) UpdateHashRate(hps float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.HashesPerSecond = hps
}
