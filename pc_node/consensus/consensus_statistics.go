package consensus

import (
	"sync"
	"time"
)

type ConsensusStatistics struct {
	BlocksAccepted        uint64
	BlocksRejected        uint64
	ValidationErrors      uint64
	AverageValidationTime time.Duration
	CurrentHeight         uint64
	CurrentTip            string
	CurrentDifficulty     string
	LastAccepted          time.Time

	mu sync.RWMutex
}

func (s *ConsensusStatistics) Snapshot() ConsensusStatistics {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return ConsensusStatistics{
		BlocksAccepted:        s.BlocksAccepted,
		BlocksRejected:        s.BlocksRejected,
		ValidationErrors:      s.ValidationErrors,
		AverageValidationTime: s.AverageValidationTime,
		CurrentHeight:         s.CurrentHeight,
		CurrentTip:            s.CurrentTip,
		CurrentDifficulty:     s.CurrentDifficulty,
		LastAccepted:          s.LastAccepted,
	}
}

func (s *ConsensusStatistics) IncBlocksAccepted(height uint64, tip string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.BlocksAccepted++
	s.CurrentHeight = height
	s.CurrentTip = tip
	s.LastAccepted = time.Now()
}

func (s *ConsensusStatistics) IncBlocksRejected() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.BlocksRejected++
}

func (s *ConsensusStatistics) IncValidationErrors() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ValidationErrors++
}

func (s *ConsensusStatistics) RecordValidationTime(d time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.AverageValidationTime == 0 {
		s.AverageValidationTime = d
	} else {
		s.AverageValidationTime = (s.AverageValidationTime + d) / 2
	}
}

func (s *ConsensusStatistics) SetDifficulty(diff string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.CurrentDifficulty = diff
}
