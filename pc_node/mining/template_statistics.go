package mining

import (
	"sync"
	"time"
)

type TemplateStatistics struct {
	TemplatesCreated     uint64
	TemplatesUpdated     uint64
	TransactionsSelected uint64
	TransactionsRejected uint64
	AverageBuildTime     time.Duration
	AverageFees          uint64
	AverageWeight        uint64
	CacheHits            uint64
	CacheMisses          uint64
	ValidationFailures   uint64

	mu sync.RWMutex
}

func (s *TemplateStatistics) Snapshot() TemplateStatistics {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return TemplateStatistics{
		TemplatesCreated:     s.TemplatesCreated,
		TemplatesUpdated:     s.TemplatesUpdated,
		TransactionsSelected: s.TransactionsSelected,
		TransactionsRejected: s.TransactionsRejected,
		AverageBuildTime:     s.AverageBuildTime,
		AverageFees:          s.AverageFees,
		AverageWeight:        s.AverageWeight,
		CacheHits:            s.CacheHits,
		CacheMisses:          s.CacheMisses,
		ValidationFailures:   s.ValidationFailures,
	}
}

func (s *TemplateStatistics) IncTemplatesCreated() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.TemplatesCreated++
}

func (s *TemplateStatistics) IncTemplatesUpdated() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.TemplatesUpdated++
}

func (s *TemplateStatistics) IncTransactionsSelected(n uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.TransactionsSelected += n
}

func (s *TemplateStatistics) IncTransactionsRejected(n uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.TransactionsRejected += n
}

func (s *TemplateStatistics) IncCacheHits() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.CacheHits++
}

func (s *TemplateStatistics) IncCacheMisses() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.CacheMisses++
}

func (s *TemplateStatistics) IncValidationFailures() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ValidationFailures++
}

func (s *TemplateStatistics) RecordBuildTime(d time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.AverageBuildTime == 0 {
		s.AverageBuildTime = d
	} else {
		s.AverageBuildTime = (s.AverageBuildTime + d) / 2
	}
}

func (s *TemplateStatistics) RecordBlockMetrics(fee uint64, weight uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.AverageFees == 0 {
		s.AverageFees = fee
	} else {
		s.AverageFees = (s.AverageFees + fee) / 2
	}

	if s.AverageWeight == 0 {
		s.AverageWeight = weight
	} else {
		s.AverageWeight = (s.AverageWeight + weight) / 2
	}
}
