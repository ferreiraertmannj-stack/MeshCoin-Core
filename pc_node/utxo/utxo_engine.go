package utxo

import (
	"context"
	"fmt"
	"time"
)

type UTXOEngine struct {
	set       *UTXOSet
	cache     *UTXOCache
	validator *TransactionValidator
	detector  *DoubleSpendDetector
	events    UTXOEvents
	stats     *UTXOStatistics
	queue     *UTXOQueue
}

func NewUTXOEngine(
	policy UTXOPolicy,
	events UTXOEvents,
	sigVal SignatureValidator,
) *UTXOEngine {
	set := NewUTXOSet()
	cache := NewUTXOCache(policy.CacheTTL)
	detector := NewDoubleSpendDetector(set)
	validator := NewTransactionValidator(policy, detector, sigVal, set, cache)
	stats := &UTXOStatistics{}

	engine := &UTXOEngine{
		set:       set,
		cache:     cache,
		validator: validator,
		detector:  detector,
		events:    events,
		stats:     stats,
	}

	queue := NewUTXOQueue(policy.QueueSize, policy.MaxWorkers, engine)
	engine.queue = queue

	return engine
}

func (e *UTXOEngine) Start(ctx context.Context) {
	e.queue.Start(ctx)
}

func (e *UTXOEngine) Stop() {
	e.queue.Stop()
}

func (e *UTXOEngine) AddBlock(block *Block) error {
	return e.queue.Enqueue(block, true)
}

func (e *UTXOEngine) RollbackBlock(block *Block) error {
	return e.queue.Enqueue(block, false)
}

func (e *UTXOEngine) processAddBlock(block *Block) error {
	// First validate all transactions
	for _, tx := range block.Transactions {
		// Assume coinbase is validated elsewhere or skip its inputs check
		// For simplicity, we assume tx validation applies to all here.
		// In a real system, coinbase is treated differently (no inputs).
		isCoinbase := len(tx.Inputs) == 1 && tx.Inputs[0].PreviousOutPoint.TxHash == "" // simple heuristic

		if !isCoinbase {
			start := time.Now()
			err := e.validator.Validate(&tx)
			e.stats.RecordValidationTime(time.Since(start))

			if err != nil {
				e.stats.IncTransactionsRejected()
				if e.events.OnTransactionRejected != nil {
					go e.events.OnTransactionRejected(tx.Hash, err.Error())
				}
				// If one fails, we should reject the block (transactionality)
				return fmt.Errorf("transaction %s failed validation: %w", tx.Hash, err)
			}
			e.stats.IncTransactionsValidated()
			if e.events.OnTransactionValidated != nil {
				go e.events.OnTransactionValidated(tx.Hash)
			}
		}
	}

	// Apply state changes
	for _, tx := range block.Transactions {
		// Spend inputs
		for _, in := range tx.Inputs {
			if in.PreviousOutPoint.TxHash == "" {
				continue // Coinbase input
			}
			e.set.Remove(in.PreviousOutPoint)
			e.cache.Invalidate(in.PreviousOutPoint)
			e.stats.IncUTXOSpent(1)
			if e.events.OnUTXOSpent != nil {
				go e.events.OnUTXOSpent(in.PreviousOutPoint)
			}
		}

		// Create outputs
		for i, out := range tx.Outputs {
			op := OutPoint{TxHash: tx.Hash, Index: uint32(i)}
			utxo := UTXO{
				Value:  out.Value,
				Script: out.Script,
				Height: block.Height,
			}
			e.set.Insert(op, utxo)
			e.stats.IncUTXOCreated(1)
			if e.events.OnUTXOCreated != nil {
				go e.events.OnUTXOCreated(op, utxo)
			}
		}
	}

	return nil
}

func (e *UTXOEngine) processRollbackBlock(block *Block) error {
	// Apply state changes in reverse
	// For full rollback, we'd need to know the spent UTXOs details.
	// As a simplification for the UTXO engine requirements, we just remove created outputs.
	// Restoring spent inputs requires an undo log, which is typically provided by the Blockchain.
	// For now, we will just delete the outputs this block created.

	for _, tx := range block.Transactions {
		for i := range tx.Outputs {
			op := OutPoint{TxHash: tx.Hash, Index: uint32(i)}
			e.set.Remove(op)
			e.cache.Invalidate(op)
		}
	}

	e.stats.IncRollbacksExecuted()
	if e.events.OnRollback != nil {
		go e.events.OnRollback(block.Height)
	}
	return nil
}

func (e *UTXOEngine) ValidateTransaction(tx *Transaction) error {
	start := time.Now()
	err := e.validator.Validate(tx)
	e.stats.RecordValidationTime(time.Since(start))

	if err != nil {
		e.stats.IncTransactionsRejected()
		return err
	}
	e.stats.IncTransactionsValidated()
	return nil
}

func (e *UTXOEngine) GetUTXO(op OutPoint) (UTXO, bool) {
	// Check cache
	u, ok := e.cache.Get(op)
	if ok {
		e.stats.IncCacheHits()
		return u, true
	}
	e.stats.IncCacheMisses()

	// Check set
	u, ok = e.set.Lookup(op)
	if ok {
		e.cache.Set(op, u)
	}
	return u, ok
}

func (e *UTXOEngine) HasUTXO(op OutPoint) bool {
	return e.set.Exists(op)
}

func (e *UTXOEngine) Snapshot() *UTXOSnapshot {
	snap := e.set.CreateSnapshot()
	e.stats.IncSnapshotsCreated()
	if e.events.OnSnapshotCreated != nil {
		go e.events.OnSnapshotCreated(snap.ID)
	}
	return snap
}

func (e *UTXOEngine) GetStatistics() UTXOStatistics {
	return e.stats.Snapshot()
}
