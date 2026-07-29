package mempool

import (
	"fmt"
	"time"
)

type TransactionPool struct {
	cache     *MempoolCache
	validator *MempoolValidator
	queue     *MempoolQueue
	events    MempoolEvents
	stats     *MempoolStatistics
	policy    MempoolPolicy
}

func NewTransactionPool(
	policy MempoolPolicy,
	events MempoolEvents,
	queueSize int,
	queueWorkers int,
) *TransactionPool {
	stats := &MempoolStatistics{}

	pool := &TransactionPool{
		cache:     NewMempoolCache(),
		validator: NewMempoolValidator(policy, events),
		events:    events,
		stats:     stats,
		policy:    policy,
	}

	pool.queue = NewMempoolQueue(pool, queueSize, queueWorkers)
	return pool
}

func (p *TransactionPool) Start() {
	p.queue.Start()
	// Start an internal cleaner for TTL expiration
	go p.cleanupLoop()
}

func (p *TransactionPool) Stop() {
	p.queue.Stop()
}

// AddTransaction queues a transaction for validation and insertion
func (p *TransactionPool) AddTransaction(tx Transaction) error {
	return p.queue.Enqueue(tx)
}

// processTransaction is called by Queue workers
func (p *TransactionPool) processTransaction(tx Transaction) error {
	start := time.Now()

	// Validation
	err := p.validator.Validate(tx)
	p.stats.RecordValidationTime(time.Since(start))
	if err != nil {
		p.stats.IncDropped()
		return err
	}

	hash := tx.GetHash()

	// Check duplicates
	if p.cache.Contains(hash) {
		p.stats.IncDuplicates()
		if p.events.OnDuplicateTransaction != nil {
			go p.events.OnDuplicateTransaction(hash)
		}
		return fmt.Errorf("duplicate transaction")
	}

	// Check RBF or Memory limits (very simple policy checking)
	if p.cache.Count() >= p.policy.MaxTransactions {
		// Attempt eviction
		evicted := p.evictLowestFee(tx.GetFee())
		if !evicted {
			p.stats.IncDropped()
			if p.events.OnPoolOverflow != nil {
				go p.events.OnPoolOverflow()
			}
			return fmt.Errorf("mempool full and tx fee too low for replacement")
		}
	}

	// Insert
	if p.cache.Add(tx) {
		p.stats.IncTransactions(tx.GetSize(), tx.GetFee())
		if p.events.OnTransactionAdded != nil {
			go p.events.OnTransactionAdded(hash)
		}
		return nil
	}

	return fmt.Errorf("failed to add")
}

// evictLowestFee attempts to evict the lowest fee transaction to make room.
// Returns true if a transaction was evicted or there was already room.
func (p *TransactionPool) evictLowestFee(newFee uint64) bool {
	// In a complete implementation, an ordered structure by Fee (like a min-heap) is used.
	// Since we are decoupled, we just perform a linear scan or assume rejection.
	// This maintains O(1) for our other lookups, but eviction is O(N).
	if !p.policy.AllowRBF {
		return false
	}

	all := p.cache.GetAll()
	var lowestTx Transaction
	var lowestFee uint64 = ^uint64(0)

	for _, tx := range all {
		if tx.GetFee() < lowestFee {
			lowestFee = tx.GetFee()
			lowestTx = tx
		}
	}

	if lowestTx != nil && newFee > lowestFee {
		p.RemoveTransaction(lowestTx.GetHash())
		return true
	}

	return false
}

func (p *TransactionPool) RemoveTransaction(hash string) {
	tx, removed := p.cache.Remove(hash)
	if removed {
		lifetime := time.Since(p.cache.GetAddedAt(hash))
		p.stats.DecTransactions(tx.GetSize(), lifetime)
		if p.events.OnTransactionRemoved != nil {
			go p.events.OnTransactionRemoved(hash)
		}
	}
}

func (p *TransactionPool) Contains(hash string) bool {
	exists := p.cache.Contains(hash)
	if exists {
		p.stats.IncHits()
	} else {
		p.stats.IncMisses()
	}
	return exists
}

func (p *TransactionPool) Get(hash string) (Transaction, bool) {
	tx, exists := p.cache.Get(hash)
	if exists {
		p.stats.IncHits()
	} else {
		p.stats.IncMisses()
	}
	return tx, exists
}

func (p *TransactionPool) Count() int {
	return p.cache.Count()
}

func (p *TransactionPool) Clear() {
	p.cache.Clear()
}

func (p *TransactionPool) Snapshot() []Transaction {
	return p.cache.GetAll()
}

func (p *TransactionPool) Statistics() MempoolStatistics {
	return p.stats.Snapshot()
}

func (p *TransactionPool) cleanupLoop() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-p.queue.ctx.Done():
			return
		case <-ticker.C:
			p.evictExpired()
		}
	}
}

func (p *TransactionPool) evictExpired() {
	all := p.cache.GetAll()
	now := time.Now()

	for _, tx := range all {
		addedAt := p.cache.GetAddedAt(tx.GetHash())
		if now.Sub(addedAt) > p.policy.TTL {
			p.RemoveTransaction(tx.GetHash())
			p.stats.IncExpired()
			if p.events.OnTransactionExpired != nil {
				go p.events.OnTransactionExpired(tx.GetHash())
			}
		}
	}
}
