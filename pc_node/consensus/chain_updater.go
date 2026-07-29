package consensus

import "fmt"

type BlockchainAppender interface {
	AppendBlock(block *Block) error
	GetHighestBlockHeight() uint64
}

type ChainUpdater struct {
	appender BlockchainAppender
	events   ConsensusEvents
}

func NewChainUpdater(appender BlockchainAppender, events ConsensusEvents) *ChainUpdater {
	return &ChainUpdater{
		appender: appender,
		events:   events,
	}
}

func (u *ChainUpdater) Update(block *Block) error {
	err := u.appender.AppendBlock(block)
	if err != nil {
		return fmt.Errorf("failed to append block to chain: %w", err)
	}

	newHeight := u.appender.GetHighestBlockHeight()

	if u.events.OnChainUpdated != nil {
		go u.events.OnChainUpdated(newHeight)
	}

	if u.events.OnTipChanged != nil {
		go u.events.OnTipChanged(block.Hash())
	}

	return nil
}
