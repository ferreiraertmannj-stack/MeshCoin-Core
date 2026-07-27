package jsonstorage

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"pc_node/storage"
	"sync"
)

type JSONEngine struct {
	filePath string
	mu       sync.RWMutex
	chain    []json.RawMessage
}

func MarshalBlock(block interface{}) ([]byte, error) {
	return json.Marshal(block)
}

func UnmarshalBlock(data []byte, block interface{}) error {
	return json.Unmarshal(data, block)
}

func NewJSONEngine() *JSONEngine {
	return &JSONEngine{}
}

func (e *JSONEngine) Open(connectionString string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.filePath = connectionString
	file, err := ioutil.ReadFile(e.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return storage.ErrNotFound
		}
		return err
	}

	if err := json.Unmarshal(file, &e.chain); err != nil {
		return storage.ErrCorruptedData
	}
	return nil
}

func (e *JSONEngine) Close() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.chain = nil
	return nil
}

func (e *JSONEngine) GetBlockByIndex(index uint64) ([]byte, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	if int(index) >= len(e.chain) {
		return nil, storage.ErrNotFound
	}
	return e.chain[index], nil
}

func (e *JSONEngine) GetLatestBlock() ([]byte, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	if len(e.chain) == 0 {
		return nil, storage.ErrNotFound
	}
	return e.chain[len(e.chain)-1], nil
}

func (e *JSONEngine) GetBalance(address string) (float64, error) {
	return 0, storage.ErrUnsupported
}

func (e *JSONEngine) NewBlockIterator() storage.Iterator {
	e.mu.RLock()
	chainCopy := make([]json.RawMessage, len(e.chain))
	copy(chainCopy, e.chain)
	e.mu.RUnlock()
	return &JSONIterator{chain: chainCopy, pos: -1}
}

func (e *JSONEngine) CreateSnapshot(path string) error {
	return storage.ErrUnsupported
}

func (e *JSONEngine) NewBatch() storage.Batch {
	return &JSONBatch{
		engine: e,
	}
}

type JSONBatch struct {
	engine       *JSONEngine
	pendingBlock []byte
}

func (b *JSONBatch) PutBlock(index uint64, blockData []byte) error {
	b.pendingBlock = blockData
	return nil
}

func (b *JSONBatch) PutBalance(address string, balance float64) error {
	return storage.ErrUnsupported
}

func (b *JSONBatch) Commit() error {
	b.engine.mu.Lock()
	defer b.engine.mu.Unlock()

	if b.pendingBlock != nil {
		b.engine.chain = append(b.engine.chain, b.pendingBlock)
	}

	data, err := json.MarshalIndent(b.engine.chain, "", "  ")
	if err != nil {
		return err
	}

	tmpFile, err := os.CreateTemp(".", "ledger_tmp_*.json")
	if err != nil {
		return err
	}
	tmpName := tmpFile.Name()

	if _, err := tmpFile.Write(data); err != nil {
		tmpFile.Close()
		os.Remove(tmpName)
		return err
	}

	if err := tmpFile.Sync(); err != nil {
		tmpFile.Close()
		os.Remove(tmpName)
		return err
	}

	tmpFile.Close()

	if err := os.Rename(tmpName, b.engine.filePath); err != nil {
		os.Remove(tmpName)
		return err
	}

	return nil
}

func (b *JSONBatch) Discard() {
	b.pendingBlock = nil
}

type JSONIterator struct {
	chain []json.RawMessage
	pos   int
}

func (it *JSONIterator) Next() bool {
	it.pos++
	return it.pos < len(it.chain)
}

func (it *JSONIterator) Value() []byte {
	if it.pos >= 0 && it.pos < len(it.chain) {
		return it.chain[it.pos]
	}
	return nil
}

func (it *JSONIterator) Error() error {
	return nil
}

func (it *JSONIterator) Close() {
	it.chain = nil
}
