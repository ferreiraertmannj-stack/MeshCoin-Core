package mempool

import (
	"sync"
	"time"
)

type TransactionNetworkStatistics struct {
	Received               uint64
	Accepted               uint64
	Rejected               uint64
	Duplicates             uint64
	Announcements          uint64
	Downloads              uint64
	Uploads                uint64
	AverageValidationTime  time.Duration
	AveragePropagationTime time.Duration
	QueueOverflows         uint64
	Retries                uint64

	mu sync.RWMutex
}

func (s *TransactionNetworkStatistics) IncReceived() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Received++
}

func (s *TransactionNetworkStatistics) IncAccepted() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Accepted++
}

func (s *TransactionNetworkStatistics) IncRejected() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Rejected++
}

func (s *TransactionNetworkStatistics) IncDuplicates() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Duplicates++
}

func (s *TransactionNetworkStatistics) IncAnnouncements() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Announcements++
}

func (s *TransactionNetworkStatistics) IncDownloads() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Downloads++
}

func (s *TransactionNetworkStatistics) IncUploads() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Uploads++
}

func (s *TransactionNetworkStatistics) IncQueueOverflows() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.QueueOverflows++
}

func (s *TransactionNetworkStatistics) IncRetries() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Retries++
}

func (s *TransactionNetworkStatistics) RecordValidationTime(d time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.AverageValidationTime == 0 {
		s.AverageValidationTime = d
	} else {
		s.AverageValidationTime = (s.AverageValidationTime + d) / 2
	}
}

func (s *TransactionNetworkStatistics) RecordPropagationTime(d time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.AveragePropagationTime == 0 {
		s.AveragePropagationTime = d
	} else {
		s.AveragePropagationTime = (s.AveragePropagationTime + d) / 2
	}
}

func (s *TransactionNetworkStatistics) Snapshot() TransactionNetworkStatistics {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return TransactionNetworkStatistics{
		Received:               s.Received,
		Accepted:               s.Accepted,
		Rejected:               s.Rejected,
		Duplicates:             s.Duplicates,
		Announcements:          s.Announcements,
		Downloads:              s.Downloads,
		Uploads:                s.Uploads,
		AverageValidationTime:  s.AverageValidationTime,
		AveragePropagationTime: s.AveragePropagationTime,
		QueueOverflows:         s.QueueOverflows,
		Retries:                s.Retries,
	}
}
