package p2p

import (
	"sync"
	"time"
)

type InventoryStatistics struct {
	InventoriesReceived uint64
	InventoriesSent     uint64
	ObjectsRequested    uint64
	ObjectsReceived     uint64
	ObjectsDelivered    uint64
	ObjectsIgnored      uint64
	ObjectsCached       uint64
	CacheHits           uint64
	CacheMisses         uint64
	Timeouts            uint64
	QueueOverflows      uint64
	Retries             uint64
	AverageRequestTime  time.Duration
	RunningWorkers      uint64
	PendingRequests     uint64
	StartTime           time.Time
	LastSynchronization time.Time

	mu sync.RWMutex
}

func (s *InventoryStatistics) Snapshot() InventoryStatistics {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return InventoryStatistics{
		InventoriesReceived: s.InventoriesReceived,
		InventoriesSent:     s.InventoriesSent,
		ObjectsRequested:    s.ObjectsRequested,
		ObjectsReceived:     s.ObjectsReceived,
		ObjectsDelivered:    s.ObjectsDelivered,
		ObjectsIgnored:      s.ObjectsIgnored,
		ObjectsCached:       s.ObjectsCached,
		CacheHits:           s.CacheHits,
		CacheMisses:         s.CacheMisses,
		Timeouts:            s.Timeouts,
		QueueOverflows:      s.QueueOverflows,
		Retries:             s.Retries,
		AverageRequestTime:  s.AverageRequestTime,
		RunningWorkers:      s.RunningWorkers,
		PendingRequests:     s.PendingRequests,
		StartTime:           s.StartTime,
		LastSynchronization: s.LastSynchronization,
	}
}

func (s *InventoryStatistics) IncInventoriesReceived() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.InventoriesReceived++
}

func (s *InventoryStatistics) IncInventoriesSent() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.InventoriesSent++
}

func (s *InventoryStatistics) IncObjectsRequested() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ObjectsRequested++
}

func (s *InventoryStatistics) IncObjectsReceived(latency time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.ObjectsReceived == 0 {
		s.AverageRequestTime = latency
	} else {
		total := int64(s.AverageRequestTime) * int64(s.ObjectsReceived)
		s.AverageRequestTime = time.Duration((total + int64(latency)) / int64(s.ObjectsReceived+1))
	}
	s.ObjectsReceived++
	s.LastSynchronization = time.Now()
}

func (s *InventoryStatistics) IncObjectsDelivered() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ObjectsDelivered++
}

func (s *InventoryStatistics) IncObjectsIgnored() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ObjectsIgnored++
}

func (s *InventoryStatistics) IncObjectsCached() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ObjectsCached++
}

func (s *InventoryStatistics) IncCacheHits() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.CacheHits++
}

func (s *InventoryStatistics) IncCacheMisses() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.CacheMisses++
}

func (s *InventoryStatistics) IncTimeouts() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Timeouts++
}

func (s *InventoryStatistics) IncQueueOverflows() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.QueueOverflows++
}

func (s *InventoryStatistics) IncRetries() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Retries++
}

func (s *InventoryStatistics) AddWorker() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.RunningWorkers++
}

func (s *InventoryStatistics) RemoveWorker() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.RunningWorkers--
}

func (s *InventoryStatistics) AddPendingRequest() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.PendingRequests++
}

func (s *InventoryStatistics) RemovePendingRequest() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.PendingRequests--
}
