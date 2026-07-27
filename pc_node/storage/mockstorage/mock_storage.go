package mockstorage

import (
	"pc_node/storage"
	"sync"
)

// MockEngine is an in-memory implementation of storage.Engine
type MockEngine struct {
	mu       sync.RWMutex
	blocks   [][]byte
	balances map[string]float64
	isOpen   bool
}

func NewMockEngine() *MockEngine {
	return &MockEngine{
		blocks:   make([][]byte, 0),
		balances: make(map[string]float64),
	}
}

// Lifecycle
func (e *MockEngine) Open(connectionString string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.isOpen = true
	return nil
}

func (e *MockEngine) Close() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if !e.isOpen {
		return storage.ErrClosed
	}
	e.isOpen = false
	return nil
}

// Block Read Operations
func (e *MockEngine) GetBlockByIndex(index uint64) ([]byte, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if !e.isOpen {
		return nil, storage.ErrClosed
	}

	if int(index) >= len(e.blocks) {
		return nil, storage.ErrNotFound
	}

	b := e.blocks[index]
	if b == nil {
		return nil, storage.ErrNotFound
	}

	// Retorna uma cópia defensiva para evitar Data Race em leituras posteriores
	c := make([]byte, len(b))
	copy(c, b)
	return c, nil
}

func (e *MockEngine) GetLatestBlock() ([]byte, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if !e.isOpen {
		return nil, storage.ErrClosed
	}

	if len(e.blocks) == 0 {
		return nil, storage.ErrNotFound
	}

	b := e.blocks[len(e.blocks)-1]
	if b == nil {
		return nil, storage.ErrNotFound
	}

	c := make([]byte, len(b))
	copy(c, b)
	return c, nil
}

// State Read Operations
func (e *MockEngine) GetBalance(address string) (float64, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if !e.isOpen {
		return 0, storage.ErrClosed
	}

	bal, ok := e.balances[address]
	if !ok {
		return 0, storage.ErrNotFound
	}
	return bal, nil
}

// Batch Operations
func (e *MockEngine) NewBatch() storage.Batch {
	return &MockBatch{
		engine:   e,
		blocks:   make(map[uint64][]byte),
		balances: make(map[string]float64),
	}
}

// Iteration
func (e *MockEngine) NewBlockIterator() storage.Iterator {
	e.mu.RLock()
	defer e.mu.RUnlock()

	// Criação de Snapshot isolado em memória (Deep Copy)
	snapshot := make([][]byte, 0, len(e.blocks))
	for _, b := range e.blocks {
		if b != nil {
			c := make([]byte, len(b))
			copy(c, b)
			snapshot = append(snapshot, c)
		}
	}

	return &MockIterator{
		blocks: snapshot,
		pos:    -1,
	}
}

// Snapshots
func (e *MockEngine) CreateSnapshot(path string) error {
	// Apenas simula a criação sem escrever em disco (in-memory)
	e.mu.RLock()
	defer e.mu.RUnlock()

	if !e.isOpen {
		return storage.ErrClosed
	}
	return nil
}

// MockBatch implements storage.Batch
type MockBatch struct {
	engine   *MockEngine
	blocks   map[uint64][]byte
	balances map[string]float64
}

func (b *MockBatch) PutBlock(index uint64, blockData []byte) error {
	c := make([]byte, len(blockData))
	copy(c, blockData)
	b.blocks[index] = c
	return nil
}

func (b *MockBatch) PutBalance(address string, balance float64) error {
	b.balances[address] = balance
	return nil
}

func (b *MockBatch) Commit() error {
	b.engine.mu.Lock()
	defer b.engine.mu.Unlock()

	if !b.engine.isOpen {
		return storage.ErrClosed
	}

	// Aplica blocos (Garante expansão da slice se índice for superior ao tamanho atual)
	for index, data := range b.blocks {
		if int(index) >= len(b.engine.blocks) {
			newBlocks := make([][]byte, int(index)+1)
			copy(newBlocks, b.engine.blocks)
			b.engine.blocks = newBlocks
		}
		b.engine.blocks[index] = data
	}

	// Aplica saldos
	for address, bal := range b.balances {
		b.engine.balances[address] = bal
	}

	// Limpa estado interno do Batch
	b.Discard()

	return nil
}

func (b *MockBatch) Discard() {
	b.blocks = make(map[uint64][]byte)
	b.balances = make(map[string]float64)
}

// MockIterator implements storage.Iterator
type MockIterator struct {
	blocks [][]byte
	pos    int
}

func (it *MockIterator) Next() bool {
	it.pos++
	return it.pos < len(it.blocks)
}

func (it *MockIterator) Value() []byte {
	if it.pos >= 0 && it.pos < len(it.blocks) {
		return it.blocks[it.pos]
	}
	return nil
}

func (it *MockIterator) Error() error {
	return nil
}

func (it *MockIterator) Close() {
	it.blocks = nil
}
