package mempool

import (
	"time"
)

type TransactionValidationPipeline struct {
	pool       *TransactionPool
	dedup      *TransactionDedup
	propagator *TransactionPropagation
	stats      *TransactionNetworkStatistics
	events     TransactionNetworkEvents
}

func NewTransactionValidationPipeline(
	pool *TransactionPool,
	dedup *TransactionDedup,
	propagator *TransactionPropagation,
	stats *TransactionNetworkStatistics,
	events TransactionNetworkEvents,
) *TransactionValidationPipeline {
	return &TransactionValidationPipeline{
		pool:       pool,
		dedup:      dedup,
		propagator: propagator,
		stats:      stats,
		events:     events,
	}
}

// Process runs the transaction through the validation stages.
// 1. Dedup (O(1) cache)
// 2. Validation / Fee Check / TTL Check (Handled by Mempool core Validator)
// 3. Mempool Insertion
// 4. Propagation (Inventory -> Gossip)
func (p *TransactionValidationPipeline) Process(tx *MsgTransaction) {
	start := time.Now()
	hash := tx.GetHash()

	// 1. Dedup (Thread-safe check and set)
	if !p.dedup.CheckAndAdd(hash) || p.pool.Contains(hash) {
		p.stats.IncDuplicates()
		if p.events.OnDuplicateTransaction != nil {
			go p.events.OnDuplicateTransaction(hash)
		}
		return
	}

	// 2 & 3. Validation and Mempool Insertion
	err := p.pool.AddTransaction(tx)
	if err != nil {
		p.stats.IncRejected()
		if p.events.OnTransactionRejected != nil {
			go p.events.OnTransactionRejected(hash, err.Error())
		}
		return
	}

	// Wait, pool.AddTransaction goes into the pool's internal Queue.
	// But in this network integration, we want to know when it's ACTUALLY accepted.
	// We can hook into OnTransactionAdded from the pool to trigger propagation!
	// Yes, `network_bridge` will register `OnTransactionAdded` in MempoolEvents
	// to call `PropagateNewTransaction`.

	// However, the prompt says "Pipeline: Receive -> Dedup -> Validation -> Fee Check -> TTL Check -> Queue -> Mempool -> Inventory -> Gossip".
	// The `pool.AddTransaction(tx)` goes to `MempoolQueue`. Then `MempoolQueue` worker does validation and adds to `MempoolCache`.
	// Since `network_bridge` has its own queue, maybe we don't need two queues?
	// But `Mempool Engine` (Fase 41) already has `queue`. The network queue just feeds `Mempool`.
	// Actually, `pool.AddTransaction(tx)` returns immediately (or error if full).

	p.stats.IncAccepted()
	p.stats.RecordValidationTime(time.Since(start))
}

// HandleAnnouncement is called when an announcement (hash) is received via Inventory.
func (p *TransactionValidationPipeline) HandleAnnouncement(peerID string, hash string) {
	p.stats.IncAnnouncements()

	if p.events.OnAnnouncementReceived != nil {
		go p.events.OnAnnouncementReceived(hash)
	}

	// 1. Check dedup
	if p.dedup.Contains(hash) {
		return
	}

	// 2. Check mempool
	if p.pool.Contains(hash) {
		return
	}

	// 3. Request
	p.stats.IncDownloads()
	err := p.propagator.RequestTransaction(peerID, hash)
	if err != nil {
		// Log or handle retry
		p.stats.IncRetries()
	}
}
