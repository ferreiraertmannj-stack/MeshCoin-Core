package mining

import "time"

// Transaction defines the generic transaction interface for the assembler
type Transaction interface {
	GetHash() string
	GetFee() uint64
	GetSize() uint64
	GetSender() string
	GetTimestamp() int64
}

// BlockTemplate defines the structure of a candidate block ready for mining
type BlockTemplate struct {
	Height       uint64
	PreviousHash string
	Transactions []Transaction
	Coinbase     Transaction
	Timestamp    int64
	Target       string // Difficulty target
	TotalFee     uint64
	TotalWeight  uint64
	MerkleRoot   string
	Version      uint32
}

// BlockchainProvider provides ledger information
type BlockchainProvider interface {
	GetHighestBlockHeight() uint64
	GetHighestBlockHash() string
	GetDifficultyTarget() string
}

// MempoolProvider provides access to pending transactions
type MempoolProvider interface {
	Snapshot() []Transaction
}

// ConsensusProvider validates final assembly boundaries
type ConsensusProvider interface {
	CalculateMerkleRoot(hashes []string) string
	ValidateBlockTemplate(template *BlockTemplate) error
}

// NetworkProvider interacts with network constraints
type NetworkProvider interface {
	GetNetworkTimestamp() int64
}

// BlockTemplateEngine orchestrates the creation of block candidates
type BlockTemplateEngine struct {
	blockchain BlockchainProvider
	mempool    MempoolProvider
	consensus  ConsensusProvider
	network    NetworkProvider

	policy    TemplatePolicy
	stats     *TemplateStatistics
	events    TemplateEvents
	cache     *TemplateCache
	assembler *BlockAssembler
	scheduler *TemplateScheduler
}

func NewBlockTemplateEngine(
	blockchain BlockchainProvider,
	mempool MempoolProvider,
	consensus ConsensusProvider,
	network NetworkProvider,
	policy TemplatePolicy,
	events TemplateEvents,
) *BlockTemplateEngine {

	stats := &TemplateStatistics{}
	cache := NewTemplateCache(policy.CacheTTL)

	feeCalc := NewFeeCalculator()
	weightCalc := NewWeightCalculator(policy)
	selector := NewTransactionSelector(weightCalc, feeCalc)
	coinbaseBuilder := NewCoinbaseBuilder(policy)
	validator := NewTemplateValidator(policy, consensus, events, stats)

	assembler := NewBlockAssembler(
		blockchain,
		mempool,
		network,
		selector,
		coinbaseBuilder,
		validator,
		consensus,
		events,
		stats,
	)

	engine := &BlockTemplateEngine{
		blockchain: blockchain,
		mempool:    mempool,
		consensus:  consensus,
		network:    network,
		policy:     policy,
		stats:      stats,
		events:     events,
		cache:      cache,
		assembler:  assembler,
	}

	queue := NewTemplateQueue(100) // Buffer for incoming triggers
	engine.scheduler = NewTemplateScheduler(engine, queue, policy.RefreshInterval)

	return engine
}

func (e *BlockTemplateEngine) Start() {
	e.scheduler.Start()
}

func (e *BlockTemplateEngine) Stop() {
	e.scheduler.Stop()
}

// GetLatestTemplate returns a cached template or builds a new one if stale
func (e *BlockTemplateEngine) GetLatestTemplate() (*BlockTemplate, error) {
	if tmpl, valid := e.cache.Get(); valid {
		e.stats.IncCacheHits()
		return tmpl, nil
	}

	e.stats.IncCacheMisses()

	// Synchronous build if no cache
	return e.BuildNewTemplate()
}

// BuildNewTemplate forces the creation of a new candidate block
func (e *BlockTemplateEngine) BuildNewTemplate() (*BlockTemplate, error) {
	start := time.Now()

	tmpl, err := e.assembler.Assemble()
	if err != nil {
		if e.events.OnTemplateRejected != nil {
			go e.events.OnTemplateRejected(err.Error())
		}
		return nil, err
	}

	e.cache.Set(tmpl)
	e.stats.IncTemplatesCreated()
	e.stats.RecordBuildTime(time.Since(start))
	e.stats.RecordBlockMetrics(tmpl.TotalFee, tmpl.TotalWeight)

	if e.events.OnTemplateCreated != nil {
		go e.events.OnTemplateCreated(tmpl.Height)
	}

	return tmpl, nil
}
