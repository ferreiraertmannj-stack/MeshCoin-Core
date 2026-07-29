package p2p

import (
	"sync"
	"time"
)

type BlockchainSyncStatistics struct {
	HeadersReceived  uint64
	HeadersValidated uint64
	BlocksRequested  uint64
	BlocksReceived   uint64
	BlocksImported   uint64
	PendingBlocks    uint64
	OrphanBlocks     uint64
	ForksDetected    uint64
	Reorganizations  uint64
	AverageSyncSpeed time.Duration // Time per block or header
	RemainingBlocks  uint64
	ETA              time.Duration

	mu sync.RWMutex
}

func (s *BlockchainSyncStatistics) Snapshot() BlockchainSyncStatistics {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return BlockchainSyncStatistics{
		HeadersReceived:  s.HeadersReceived,
		HeadersValidated: s.HeadersValidated,
		BlocksRequested:  s.BlocksRequested,
		BlocksReceived:   s.BlocksReceived,
		BlocksImported:   s.BlocksImported,
		PendingBlocks:    s.PendingBlocks,
		OrphanBlocks:     s.OrphanBlocks,
		ForksDetected:    s.ForksDetected,
		Reorganizations:  s.Reorganizations,
		AverageSyncSpeed: s.AverageSyncSpeed,
		RemainingBlocks:  s.RemainingBlocks,
		ETA:              s.ETA,
	}
}

func (s *BlockchainSyncStatistics) IncHeadersReceived(n int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.HeadersReceived += uint64(n)
}

func (s *BlockchainSyncStatistics) IncHeadersValidated(n int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.HeadersValidated += uint64(n)
}

func (s *BlockchainSyncStatistics) IncBlocksRequested(n int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.BlocksRequested += uint64(n)
}

func (s *BlockchainSyncStatistics) IncBlocksReceived(n int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.BlocksReceived += uint64(n)
}

func (s *BlockchainSyncStatistics) IncBlocksImported(n int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.BlocksImported += uint64(n)
}

func (s *BlockchainSyncStatistics) SetPendingBlocks(n int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.PendingBlocks = uint64(n)
}

func (s *BlockchainSyncStatistics) SetOrphanBlocks(n int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.OrphanBlocks = uint64(n)
}

func (s *BlockchainSyncStatistics) IncForksDetected() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ForksDetected++
}

func (s *BlockchainSyncStatistics) IncReorganizations() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Reorganizations++
}

func (s *BlockchainSyncStatistics) UpdateRemaining(n int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.RemainingBlocks = uint64(n)
}
