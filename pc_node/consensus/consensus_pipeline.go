package consensus

import (
	"fmt"
	"time"
)

type NetworkProvider interface {
	GetNetworkTimestamp() int64
}

// BlockAcceptancePipeline is the entry point for blocks found by miners or received via network.
type BlockAcceptancePipeline struct {
	validator    *BlockValidator
	chainUpdater *ChainUpdater
	difficulty   DifficultyUpdater
	stats        *ConsensusStatistics
	events       ConsensusEvents
	network      NetworkProvider
	appender     BlockchainAppender
}

func NewBlockAcceptancePipeline(
	validator *BlockValidator,
	chainUpdater *ChainUpdater,
	difficulty DifficultyUpdater,
	stats *ConsensusStatistics,
	events ConsensusEvents,
	network NetworkProvider,
	appender BlockchainAppender,
) *BlockAcceptancePipeline {
	return &BlockAcceptancePipeline{
		validator:    validator,
		chainUpdater: chainUpdater,
		difficulty:   difficulty,
		stats:        stats,
		events:       events,
		network:      network,
		appender:     appender,
	}
}

func (p *BlockAcceptancePipeline) ProcessBlock(block *Block) error {
	start := time.Now()
	hash := block.Hash()

	currentNetTime := p.network.GetNetworkTimestamp()

	// 1. Validation
	err := p.validator.Validate(block, currentNetTime, block.PrevHash) // Simplified
	if err != nil {
		p.stats.IncBlocksRejected()
		if p.events.OnBlockRejected != nil {
			go p.events.OnBlockRejected(hash, err.Error())
		}
		return fmt.Errorf("block validation failed: %w", err)
	}

	// 2. Integration / Append
	err = p.chainUpdater.Update(block)
	if err != nil {
		p.stats.IncBlocksRejected()
		if p.events.OnBlockRejected != nil {
			go p.events.OnBlockRejected(hash, "Chain update failed")
		}
		return err
	}

	// 3. Difficulty Retargeting Check
	newTarget, newDiff, changed := p.difficulty.UpdateDifficulty(block.Height)
	if changed {
		p.stats.SetDifficulty(newDiff)
		if p.events.OnDifficultyChanged != nil {
			go p.events.OnDifficultyChanged("old", newTarget)
		}
	}

	// 4. Accept
	p.stats.IncBlocksAccepted(block.Height, hash)
	p.stats.RecordValidationTime(time.Since(start))

	if p.events.OnBlockAccepted != nil {
		go p.events.OnBlockAccepted(hash, block.Height)
	}

	return nil
}
