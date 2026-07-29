package runtime

import (
	"sync"
	"time"
)

type RuntimeStatistics struct {
	ModulesLoaded       uint64
	ModulesFailed       uint64
	HealthChecks        uint64
	Restarts            uint64
	Shutdowns           uint64
	AverageStartupTime  time.Duration
	AverageShutdownTime time.Duration
	RunningServices     uint64
	MemoryUsage         uint64
	Goroutines          uint64

	mu sync.RWMutex
}

func (s *RuntimeStatistics) Snapshot() RuntimeStatistics {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return RuntimeStatistics{
		ModulesLoaded:       s.ModulesLoaded,
		ModulesFailed:       s.ModulesFailed,
		HealthChecks:        s.HealthChecks,
		Restarts:            s.Restarts,
		Shutdowns:           s.Shutdowns,
		AverageStartupTime:  s.AverageStartupTime,
		AverageShutdownTime: s.AverageShutdownTime,
		RunningServices:     s.RunningServices,
		MemoryUsage:         s.MemoryUsage,
		Goroutines:          s.Goroutines,
	}
}

func (s *RuntimeStatistics) IncModulesLoaded() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ModulesLoaded++
}

func (s *RuntimeStatistics) IncModulesFailed() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ModulesFailed++
}

func (s *RuntimeStatistics) IncHealthChecks() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.HealthChecks++
}

func (s *RuntimeStatistics) IncRestarts() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Restarts++
}

func (s *RuntimeStatistics) IncShutdowns() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Shutdowns++
}

func (s *RuntimeStatistics) SetRunningServices(count uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.RunningServices = count
}

func (s *RuntimeStatistics) RecordStartupTime(d time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.AverageStartupTime == 0 {
		s.AverageStartupTime = d
	} else {
		s.AverageStartupTime = (s.AverageStartupTime + d) / 2
	}
}

func (s *RuntimeStatistics) RecordShutdownTime(d time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.AverageShutdownTime == 0 {
		s.AverageShutdownTime = d
	} else {
		s.AverageShutdownTime = (s.AverageShutdownTime + d) / 2
	}
}
