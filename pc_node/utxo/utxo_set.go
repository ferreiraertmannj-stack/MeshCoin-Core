package utxo

import (
	"sync"
)

type OutPoint struct {
	TxHash string
	Index  uint32
}

type UTXO struct {
	Value  uint64
	Script []byte
	Height uint64
}

// UTXOSet holds the map[OutPoint]UTXO protected by sync.RWMutex
type UTXOSet struct {
	utxos map[OutPoint]UTXO
	mu    sync.RWMutex
}

func NewUTXOSet() *UTXOSet {
	return &UTXOSet{
		utxos: make(map[OutPoint]UTXO),
	}
}

func (s *UTXOSet) Insert(op OutPoint, utxo UTXO) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.utxos[op] = utxo
}

func (s *UTXOSet) Remove(op OutPoint) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.utxos, op)
}

func (s *UTXOSet) Update(op OutPoint, utxo UTXO) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.utxos[op] = utxo
}

func (s *UTXOSet) Lookup(op OutPoint) (UTXO, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.utxos[op]
	return u, ok
}

func (s *UTXOSet) Exists(op OutPoint) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.utxos[op]
	return ok
}
