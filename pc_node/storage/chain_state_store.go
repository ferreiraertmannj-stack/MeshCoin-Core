package storage

import (
	"fmt"
	"sync"
)

type ChainState struct {
	Height     uint64
	BestBlock  string
	Difficulty uint64
	TotalWork  uint64
	Supply     uint64
	Timestamp  int64
	ChainWork  string
}

type ChainStateStore struct {
	mu    sync.RWMutex
	state *ChainState
}

func NewChainStateStore() *ChainStateStore {
	return &ChainStateStore{}
}

func (s *ChainStateStore) Save(state ChainState) {
	s.mu.Lock()
	defer s.mu.Unlock()
	// Deep copy
	s.state = &ChainState{
		Height:     state.Height,
		BestBlock:  state.BestBlock,
		Difficulty: state.Difficulty,
		TotalWork:  state.TotalWork,
		Supply:     state.Supply,
		Timestamp:  state.Timestamp,
		ChainWork:  state.ChainWork,
	}
}

func (s *ChainStateStore) Load() (ChainState, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.state == nil {
		return ChainState{}, fmt.Errorf("chain state not initialized")
	}
	return *s.state, nil
}
