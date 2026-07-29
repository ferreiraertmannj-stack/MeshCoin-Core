package mining

import (
	"context"
)

type BlockTemplateProvider interface {
	GetLatestTemplate() (*BlockTemplate, error)
}

type MiningDifficultyProvider interface {
	GetCurrentDifficulty() uint64
	GetCurrentTarget() string
}

type Block struct {
	Header       BlockHeader
	Transactions []Transaction
}

type BlockHeader struct {
	Version    uint32
	PrevHash   string
	MerkleRoot string
	Timestamp  int64
	Target     string
	Nonce      uint64
	ExtraNonce uint64
}

func (h *BlockHeader) Hash(pipeline HashPipeline) string {
	return pipeline.HashHeader(*h)
}

// MiningEngine orchestrates the PoW execution
type MiningEngine struct {
	templateProvider BlockTemplateProvider
	consensus        ConsensusProvider
	difficulty       MiningDifficultyProvider
	network          NetworkProvider

	policy    MiningPolicy
	stats     *MiningStatistics
	events    MiningEvents
	cache     *MiningCache
	queue     *MiningQueue
	scheduler *MiningScheduler

	hashPipeline HashPipeline
}

func NewMiningEngine(
	templateProvider BlockTemplateProvider,
	consensus ConsensusProvider,
	difficulty MiningDifficultyProvider,
	network NetworkProvider,
	policy MiningPolicy,
	events MiningEvents,
	pipeline HashPipeline,
) *MiningEngine {

	stats := &MiningStatistics{}
	cache := NewMiningCache(policy.CacheTTL)

	queue := NewMiningQueue(policy.QueueSize, policy.MaxWorkers)
	scheduler := NewMiningScheduler(queue, events, policy, stats, cache, pipeline, difficulty, network)

	return &MiningEngine{
		templateProvider: templateProvider,
		consensus:        consensus,
		difficulty:       difficulty,
		network:          network,
		policy:           policy,
		stats:            stats,
		events:           events,
		cache:            cache,
		queue:            queue,
		scheduler:        scheduler,
		hashPipeline:     pipeline,
	}
}

func (e *MiningEngine) Start(ctx context.Context) error {
	if e.events.OnMiningStarted != nil {
		go e.events.OnMiningStarted()
	}
	e.queue.Start(ctx)
	e.scheduler.Start(ctx)

	// Fetch first template and schedule
	tmpl, err := e.templateProvider.GetLatestTemplate()
	if err == nil && tmpl != nil {
		e.scheduler.ScheduleJob(tmpl)
	}

	return nil
}

func (e *MiningEngine) Stop() {
	if e.events.OnMiningStopped != nil {
		go e.events.OnMiningStopped()
	}
	e.scheduler.Stop()
	e.queue.Stop()
}

func (e *MiningEngine) SubmitTemplate(tmpl *BlockTemplate) {
	e.scheduler.ScheduleJob(tmpl)
}

func (e *MiningEngine) GetStatistics() MiningStatistics {
	return e.stats.Snapshot()
}
