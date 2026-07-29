package consensus

// BlockAcceptance API facade to hide internal pipeline complexity
type BlockAcceptance struct {
	pipeline *BlockAcceptancePipeline
}

func NewBlockAcceptance(
	policy ConsensusPolicy,
	events ConsensusEvents,
	appender BlockchainAppender,
	difficulty DifficultyUpdater,
	network NetworkProvider,
) *BlockAcceptance {

	stats := &ConsensusStatistics{}

	rewardVal := NewRewardValidator(policy)
	cbVal := NewCoinbaseValidator(rewardVal)
	validator := NewBlockValidator(policy, cbVal, difficulty)

	chainUp := NewChainUpdater(appender, events)

	pipeline := NewBlockAcceptancePipeline(
		validator,
		chainUp,
		difficulty,
		stats,
		events,
		network,
		appender,
	)

	return &BlockAcceptance{
		pipeline: pipeline,
	}
}

func (b *BlockAcceptance) Process(block *Block) error {
	return b.pipeline.ProcessBlock(block)
}

func (b *BlockAcceptance) Statistics() ConsensusStatistics {
	return b.pipeline.stats.Snapshot()
}
