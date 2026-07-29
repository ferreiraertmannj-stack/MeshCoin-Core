package p2p

import (
	"sync"
	"time"
)

type RouterEvents struct {
	OnMessageReceived   func(peerID string, msgType string)
	OnMessageDispatched func(peerID string, msgType string, duration time.Duration)
	OnUnknownMessage    func(peerID string, msgType string)
	OnDispatchError     func(peerID string, msgType string, err error)
	OnPeerDisconnected  func(peerID string)
	OnRouterStarted     func()
	OnRouterStopped     func()
}

type RouterStatistics struct {
	MessagesReceived    uint64
	MessagesSent        uint64
	MessagesDispatched  uint64
	MessagesDropped     uint64
	DispatchErrors      uint64
	UnknownMessages     uint64
	Broadcasts          uint64
	RunningHandlers     uint64
	AverageDispatchTime time.Duration
	StartTime           time.Time
	LastMessageTime     time.Time

	mu sync.RWMutex
}

func (s *RouterStatistics) AddReceived() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.MessagesReceived++
	s.LastMessageTime = time.Now()
}

func (s *RouterStatistics) AddSent() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.MessagesSent++
}

func (s *RouterStatistics) AddDispatched(duration time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Update moving average (simple)
	if s.MessagesDispatched == 0 {
		s.AverageDispatchTime = duration
	} else {
		// (avg * count + new) / (count + 1)
		total := int64(s.AverageDispatchTime) * int64(s.MessagesDispatched)
		s.AverageDispatchTime = time.Duration((total + int64(duration)) / int64(s.MessagesDispatched+1))
	}
	s.MessagesDispatched++
}

func (s *RouterStatistics) AddDropped() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.MessagesDropped++
}

func (s *RouterStatistics) AddDispatchError() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.DispatchErrors++
}

func (s *RouterStatistics) AddUnknown() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.UnknownMessages++
}

func (s *RouterStatistics) AddBroadcast() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Broadcasts++
}

func (s *RouterStatistics) IncRunning() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.RunningHandlers++
}

func (s *RouterStatistics) DecRunning() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.RunningHandlers--
}

func (s *RouterStatistics) Snapshot() RouterStatistics {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return RouterStatistics{
		MessagesReceived:    s.MessagesReceived,
		MessagesSent:        s.MessagesSent,
		MessagesDispatched:  s.MessagesDispatched,
		MessagesDropped:     s.MessagesDropped,
		DispatchErrors:      s.DispatchErrors,
		UnknownMessages:     s.UnknownMessages,
		Broadcasts:          s.Broadcasts,
		RunningHandlers:     s.RunningHandlers,
		AverageDispatchTime: s.AverageDispatchTime,
		StartTime:           s.StartTime,
		LastMessageTime:     s.LastMessageTime,
	}
}
