package storage

import (
	"context"
	"time"
)

type StorageEngine struct {
	policy      StoragePolicy
	events      StorageEvents
	stats       *StorageStatistics
	cache       *StorageCache
	blockStore  *BlockStore
	txStore     *TransactionStore
	utxoStore   *UTXOStore
	chainStore  *ChainStateStore
	metaStore   *MetadataStore
	queue       *StorageQueue
	compactor   *StorageCompactor
	snapshotter *StorageSnapshot
	recovery    *StorageRecovery
}

func NewStorageEngine(policy StoragePolicy, events StorageEvents) *StorageEngine {
	stats := &StorageStatistics{}
	cache := NewStorageCache(policy.CacheSize, policy.CacheTTL)

	engine := &StorageEngine{
		policy:      policy,
		events:      events,
		stats:       stats,
		cache:       cache,
		blockStore:  NewBlockStore(stats, cache),
		txStore:     NewTransactionStore(stats, cache),
		utxoStore:   NewUTXOStore(stats, cache),
		chainStore:  NewChainStateStore(),
		metaStore:   NewMetadataStore(),
		compactor:   NewStorageCompactor(stats),
		snapshotter: NewStorageSnapshot(stats),
	}

	engine.recovery = NewStorageRecovery(engine.chainStore)
	engine.queue = NewStorageQueue(policy.QueueSize, 4, engine) // 4 workers

	return engine
}

func (e *StorageEngine) Start(ctx context.Context) error {
	e.queue.Start(ctx)
	// Try recover
	e.recovery.Recover()
	return nil
}

func (e *StorageEngine) Close() {
	e.queue.Stop()
}

func (e *StorageEngine) enqueue(action string, data interface{}) error {
	start := time.Now()

	job := &StorageJob{
		Action: action,
		Data:   data,
		Done:   make(chan error, 1),
	}

	err := e.queue.Enqueue(job)

	e.stats.RecordLatency(time.Since(start))

	if err != nil && e.events.OnStorageError != nil {
		go e.events.OnStorageError(err)
	}

	return err
}

func (e *StorageEngine) SaveBlock(b *Block) error {
	err := e.enqueue("SaveBlock", b)
	if err == nil && e.events.OnBlockSaved != nil {
		go e.events.OnBlockSaved(b.Hash)
	}
	return err
}

func (e *StorageEngine) LoadBlock(hash string) (*Block, error) {
	return e.blockStore.LoadByHash(hash)
}

func (e *StorageEngine) SaveTransaction(tx *Transaction) error {
	err := e.enqueue("SaveTransaction", tx)
	if err == nil && e.events.OnTransactionSaved != nil {
		go e.events.OnTransactionSaved(tx.Hash)
	}
	return err
}

func (e *StorageEngine) LoadTransaction(hash string) (*Transaction, error) {
	return e.txStore.Load(hash)
}

func (e *StorageEngine) SaveUTXO(u *UTXOEntry) error {
	err := e.enqueue("SaveUTXO", u)
	if err == nil && e.events.OnUTXOUpdated != nil {
		go e.events.OnUTXOUpdated(u.Outpoint)
	}
	return err
}

func (e *StorageEngine) LoadUTXO(outpoint string) (*UTXOEntry, error) {
	return e.utxoStore.Load(outpoint)
}

func (e *StorageEngine) SaveChainState(state ChainState) error {
	e.chainStore.Save(state)
	return nil
}

func (e *StorageEngine) LoadChainState() (ChainState, error) {
	return e.chainStore.Load()
}

func (e *StorageEngine) Snapshot() (*Snapshot, error) {
	snap, err := e.snapshotter.Create()
	if err == nil && e.events.OnSnapshotCreated != nil {
		go e.events.OnSnapshotCreated(snap.ID)
	}
	return snap, err
}

func (e *StorageEngine) Compact(ctx context.Context) error {
	err := e.compactor.Compact(ctx)
	if err == nil && e.events.OnDatabaseCompacted != nil {
		go e.events.OnDatabaseCompacted()
	}
	return err
}

func (e *StorageEngine) GetStatistics() StorageStatistics {
	return e.stats.Snapshot()
}
