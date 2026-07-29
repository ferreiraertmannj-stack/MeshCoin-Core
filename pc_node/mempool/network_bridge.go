package mempool

import (
	"encoding/json"
	"time"
)

type NetworkBridge struct {
	pool       *TransactionPool
	queue      *TransactionQueue
	dedup      *TransactionDedup
	propagator *TransactionPropagation
	stats      *TransactionNetworkStatistics
	events     TransactionNetworkEvents
	pipeline   *TransactionValidationPipeline
	handlers   *TransactionHandlers
}

type NetworkBridgeConfig struct {
	QueueSize int
	Workers   int
	DedupTTL  time.Duration
}

func NewNetworkBridge(
	pool *TransactionPool,
	gossip GossipProtocol,
	inv InventoryProtocol,
	router RouterProtocol,
	routerHandler Router,
	events TransactionNetworkEvents,
	config NetworkBridgeConfig,
) *NetworkBridge {

	stats := &TransactionNetworkStatistics{}
	dedup := NewTransactionDedup(config.DedupTTL)
	propagator := NewTransactionPropagation(gossip, inv, router)

	pipeline := NewTransactionValidationPipeline(
		pool,
		dedup,
		propagator,
		stats,
		events,
	)

	queue := NewTransactionQueue(config.QueueSize, config.Workers, pipeline)

	handlers := NewTransactionHandlers(pipeline, queue, stats, events)
	if routerHandler != nil {
		handlers.Register(routerHandler)
	}

	bridge := &NetworkBridge{
		pool:       pool,
		queue:      queue,
		dedup:      dedup,
		propagator: propagator,
		stats:      stats,
		events:     events,
		pipeline:   pipeline,
		handlers:   handlers,
	}

	// Hook into Mempool Core events to trigger propagation
	pool.events.OnTransactionAdded = func(hash string) {
		tx, exists := pool.Get(hash)
		if exists {
			raw, _ := json.Marshal(tx)
			_ = propagator.PropagateNewTransaction(hash, raw)
			if events.OnTransactionPropagated != nil {
				go events.OnTransactionPropagated(hash)
			}
		}
	}

	return bridge
}

func (b *NetworkBridge) Start() {
	b.queue.Start()
}

func (b *NetworkBridge) Stop() {
	b.queue.Stop()
}

func (b *NetworkBridge) Statistics() TransactionNetworkStatistics {
	return b.stats.Snapshot()
}
