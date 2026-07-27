package badgerstorage

import (
	"encoding/binary"
	"fmt"
	"math"
	"sync"

	"github.com/dgraph-io/badger/v4"
	"pc_node/storage"
)

// BadgerEngine implementa a interface storage.Engine usando BadgerDB.
type BadgerEngine struct {
	db *badger.DB
	mu sync.RWMutex
}

func NewBadgerEngine() *BadgerEngine {
	return &BadgerEngine{}
}

func makeBlockIndexKey(index uint64) []byte {
	// block/index/ + 8 bytes uint64
	key := make([]byte, 12+8)
	copy(key, []byte("block/index/"))
	binary.BigEndian.PutUint64(key[12:], index)
	return key
}

func makeBalanceKey(address string) []byte {
	return []byte(fmt.Sprintf("balance/%s", address))
}

func (e *BadgerEngine) Open(connectionString string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	opts := badger.DefaultOptions(connectionString)
	opts.Logger = nil // Desabilitar os logs intrusivos do Badger por enquanto

	db, err := badger.Open(opts)
	if err != nil {
		return err
	}
	e.db = db
	return nil
}

func (e *BadgerEngine) Close() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.db == nil {
		return storage.ErrClosed
	}

	err := e.db.Close()
	e.db = nil
	return err
}

func (e *BadgerEngine) GetBlockByIndex(index uint64) ([]byte, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if e.db == nil {
		return nil, storage.ErrClosed
	}

	var data []byte
	err := e.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(makeBlockIndexKey(index))
		if err != nil {
			if err == badger.ErrKeyNotFound {
				return storage.ErrNotFound
			}
			return err
		}

		valCopy, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}
		data = valCopy
		return nil
	})

	return data, err
}

func (e *BadgerEngine) GetLatestBlock() ([]byte, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if e.db == nil {
		return nil, storage.ErrClosed
	}

	var data []byte
	err := e.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Reverse = true
		it := txn.NewIterator(opts)
		defer it.Close()

		prefix := []byte("block/index/")
		seekKey := append(prefix, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF)

		var found bool
		for it.Seek(seekKey); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			valCopy, err := item.ValueCopy(nil)
			if err != nil {
				return err
			}
			data = valCopy
			found = true
			break // Pega apenas o último
		}

		if !found {
			return storage.ErrNotFound
		}
		return nil
	})

	return data, err
}

func (e *BadgerEngine) GetBalance(address string) (float64, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if e.db == nil {
		return 0, storage.ErrClosed
	}

	var bal float64
	err := e.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(makeBalanceKey(address))
		if err != nil {
			if err == badger.ErrKeyNotFound {
				return storage.ErrNotFound
			}
			return err
		}

		valCopy, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}

		if len(valCopy) == 8 {
			bits := binary.BigEndian.Uint64(valCopy)
			bal = math.Float64frombits(bits)
		}
		return nil
	})

	return bal, err
}

func (e *BadgerEngine) NewBatch() storage.Batch {
	return &BadgerBatch{
		engine:   e,
		blocks:   make(map[uint64][]byte),
		balances: make(map[string]float64),
	}
}

func (e *BadgerEngine) NewBlockIterator() storage.Iterator {
	if e.db == nil {
		return &BadgerIterator{}
	}

	txn := e.db.NewTransaction(false)
	opts := badger.DefaultIteratorOptions
	opts.Reverse = false
	it := txn.NewIterator(opts)
	prefix := []byte("block/index/")
	it.Seek(prefix)

	return &BadgerIterator{
		txn:     txn,
		it:      it,
		prefix:  prefix,
		started: false,
	}
}

func (e *BadgerEngine) CreateSnapshot(path string) error {
	// Snapshots inteligentes postergados
	return nil
}

// BadgerBatch implementa storage.Batch
type BadgerBatch struct {
	engine   *BadgerEngine
	blocks   map[uint64][]byte
	balances map[string]float64
}

func (b *BadgerBatch) PutBlock(index uint64, blockData []byte) error {
	b.blocks[index] = blockData
	return nil
}

func (b *BadgerBatch) PutBalance(address string, balance float64) error {
	b.balances[address] = balance
	return nil
}

func (b *BadgerBatch) Commit() error {
	b.engine.mu.RLock()
	defer b.engine.mu.RUnlock()

	if b.engine.db == nil {
		return storage.ErrClosed
	}

	txn := b.engine.db.NewTransaction(true)
	defer txn.Discard()

	for index, data := range b.blocks {
		if err := txn.Set(makeBlockIndexKey(index), data); err != nil {
			return err
		}
	}

	for address, bal := range b.balances {
		bits := math.Float64bits(bal)
		buf := make([]byte, 8)
		binary.BigEndian.PutUint64(buf, bits)
		if err := txn.Set(makeBalanceKey(address), buf); err != nil {
			return err
		}
	}

	if err := txn.Commit(); err != nil {
		return err
	}

	b.Discard()
	return nil
}

func (b *BadgerBatch) Discard() {
	b.blocks = make(map[uint64][]byte)
	b.balances = make(map[string]float64)
}

// BadgerIterator implementa storage.Iterator
type BadgerIterator struct {
	txn     *badger.Txn
	it      *badger.Iterator
	prefix  []byte
	started bool
	currVal []byte
	err     error
}

func (it *BadgerIterator) Next() bool {
	if it.txn == nil {
		return false
	}

	if !it.started {
		it.started = true
	} else {
		it.it.Next()
	}

	if it.it.ValidForPrefix(it.prefix) {
		item := it.it.Item()
		valCopy, err := item.ValueCopy(nil)
		if err != nil {
			it.err = err
			return false
		}
		it.currVal = valCopy
		return true
	}
	return false
}

func (it *BadgerIterator) Value() []byte {
	return it.currVal
}

func (it *BadgerIterator) Error() error {
	return it.err
}

func (it *BadgerIterator) Close() {
	if it.it != nil {
		it.it.Close()
	}
	if it.txn != nil {
		it.txn.Discard()
	}
}
