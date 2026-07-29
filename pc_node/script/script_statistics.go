package script

import (
	"sync"
	"time"
)

type ScriptStatistics struct {
	ScriptsExecuted      uint64
	ScriptsFailed        uint64
	SignaturesVerified   uint64
	SignatureFailures    uint64
	OpcodeCount          uint64
	AverageExecutionTime time.Duration
	CacheHits            uint64
	CacheMisses          uint64
	QueueOverflows       uint64

	mu sync.RWMutex
}

func (s *ScriptStatistics) Snapshot() ScriptStatistics {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return ScriptStatistics{
		ScriptsExecuted:      s.ScriptsExecuted,
		ScriptsFailed:        s.ScriptsFailed,
		SignaturesVerified:   s.SignaturesVerified,
		SignatureFailures:    s.SignatureFailures,
		OpcodeCount:          s.OpcodeCount,
		AverageExecutionTime: s.AverageExecutionTime,
		CacheHits:            s.CacheHits,
		CacheMisses:          s.CacheMisses,
		QueueOverflows:       s.QueueOverflows,
	}
}

func (s *ScriptStatistics) IncScriptsExecuted() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ScriptsExecuted++
}

func (s *ScriptStatistics) IncScriptsFailed() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ScriptsFailed++
}

func (s *ScriptStatistics) IncSignaturesVerified() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.SignaturesVerified++
}

func (s *ScriptStatistics) IncSignatureFailures() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.SignatureFailures++
}

func (s *ScriptStatistics) IncOpcodeCount(n int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.OpcodeCount += uint64(n)
}

func (s *ScriptStatistics) IncCacheHits() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.CacheHits++
}

func (s *ScriptStatistics) IncCacheMisses() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.CacheMisses++
}

func (s *ScriptStatistics) IncQueueOverflows() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.QueueOverflows++
}

func (s *ScriptStatistics) RecordExecutionTime(d time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.AverageExecutionTime == 0 {
		s.AverageExecutionTime = d
	} else {
		s.AverageExecutionTime = (s.AverageExecutionTime + d) / 2
	}
}
