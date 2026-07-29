package p2p

import (
	"sync"
	"time"
)

type GossipStatistics struct {
	PublishedMessages      uint64
	ReceivedMessages       uint64
	ForwardedMessages      uint64
	DroppedMessages        uint64
	DuplicateMessages      uint64
	TTLExpired             uint64
	QueueOverflows         uint64
	Retries                uint64
	CacheHits              uint64
	CacheMisses            uint64
	AveragePropagationTime time.Duration
	RunningWorkers         uint64
	ActivePeers            uint64
	StartTime              time.Time
	LastPropagation        time.Time

	mu sync.RWMutex
}

func (s *GossipStatistics) Snapshot() GossipStatistics {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return GossipStatistics{
		PublishedMessages:      s.PublishedMessages,
		ReceivedMessages:       s.ReceivedMessages,
		ForwardedMessages:      s.ForwardedMessages,
		DroppedMessages:        s.DroppedMessages,
		DuplicateMessages:      s.DuplicateMessages,
		TTLExpired:             s.TTLExpired,
		QueueOverflows:         s.QueueOverflows,
		Retries:                s.Retries,
		CacheHits:              s.CacheHits,
		CacheMisses:            s.CacheMisses,
		AveragePropagationTime: s.AveragePropagationTime,
		RunningWorkers:         s.RunningWorkers,
		ActivePeers:            s.ActivePeers,
		StartTime:              s.StartTime,
		LastPropagation:        s.LastPropagation,
	}
}

func (s *GossipStatistics) IncPublished() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.PublishedMessages++
}

func (s *GossipStatistics) IncReceived() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ReceivedMessages++
}

func (s *GossipStatistics) IncForwarded(latency time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.ForwardedMessages == 0 {
		s.AveragePropagationTime = latency
	} else {
		total := int64(s.AveragePropagationTime) * int64(s.ForwardedMessages)
		s.AveragePropagationTime = time.Duration((total + int64(latency)) / int64(s.ForwardedMessages+1))
	}
	s.ForwardedMessages++
	s.LastPropagation = time.Now()
}

func (s *GossipStatistics) IncDropped() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.DroppedMessages++
}

func (s *GossipStatistics) IncDuplicate() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.DuplicateMessages++
}

func (s *GossipStatistics) IncTTLExpired() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.TTLExpired++
}

func (s *GossipStatistics) IncQueueOverflow() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.QueueOverflows++
}

func (s *GossipStatistics) IncRetries() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Retries++
}

func (s *GossipStatistics) IncCacheHits() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.CacheHits++
}

func (s *GossipStatistics) IncCacheMisses() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.CacheMisses++
}

func (s *GossipStatistics) AddWorker() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.RunningWorkers++
}

func (s *GossipStatistics) RemoveWorker() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.RunningWorkers--
}

func (s *GossipStatistics) UpdateActivePeers(count int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ActivePeers = uint64(count)
}
