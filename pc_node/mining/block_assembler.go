package mining

import "fmt"

type BlockAssembler struct {
	blockchain      BlockchainProvider
	mempool         MempoolProvider
	network         NetworkProvider
	selector        *TransactionSelector
	coinbaseBuilder *CoinbaseBuilder
	validator       *TemplateValidator
	consensus       ConsensusProvider
	events          TemplateEvents
	stats           *TemplateStatistics
}

func NewBlockAssembler(
	blockchain BlockchainProvider,
	mempool MempoolProvider,
	network NetworkProvider,
	selector *TransactionSelector,
	coinbaseBuilder *CoinbaseBuilder,
	validator *TemplateValidator,
	consensus ConsensusProvider,
	events TemplateEvents,
	stats *TemplateStatistics,
) *BlockAssembler {
	return &BlockAssembler{
		blockchain:      blockchain,
		mempool:         mempool,
		network:         network,
		selector:        selector,
		coinbaseBuilder: coinbaseBuilder,
		validator:       validator,
		consensus:       consensus,
		events:          events,
		stats:           stats,
	}
}

// Assemble pulls together mempool, blockchain tips, network time,
// selects transactions, and builds a candidate.
func (a *BlockAssembler) Assemble() (*BlockTemplate, error) {
	// 1. Gather context
	height := a.blockchain.GetHighestBlockHeight() + 1
	prevHash := a.blockchain.GetHighestBlockHash()
	target := a.blockchain.GetDifficultyTarget()
	timestamp := a.network.GetNetworkTimestamp()

	// 2. Snapshot mempool
	mempoolSnapshot := a.mempool.Snapshot()

	// 3. Select transactions
	selected, weight := a.selector.Select(mempoolSnapshot)

	// 4. Fees
	totalFees := uint64(0)
	for _, tx := range selected {
		totalFees += tx.GetFee()
		if a.events.OnTransactionSelected != nil {
			go a.events.OnTransactionSelected(tx.GetHash())
		}
	}

	a.stats.IncTransactionsSelected(uint64(len(selected)))
	a.stats.IncTransactionsRejected(uint64(len(mempoolSnapshot) - len(selected)))

	// 5. Coinbase
	coinbase := a.coinbaseBuilder.Build(height, totalFees)
	if a.events.OnCoinbaseGenerated != nil {
		go a.events.OnCoinbaseGenerated(a.coinbaseBuilder.policy.BlockReward, totalFees)
	}

	// 6. Merkle Root
	hashes := []string{coinbase.GetHash()}
	for _, tx := range selected {
		hashes = append(hashes, tx.GetHash())
	}
	merkleRoot := a.consensus.CalculateMerkleRoot(hashes)

	// 7. Compose BlockTemplate
	tmpl := &BlockTemplate{
		Height:       height,
		PreviousHash: prevHash,
		Transactions: selected,
		Coinbase:     coinbase,
		Timestamp:    timestamp,
		Target:       target,
		TotalFee:     totalFees,
		TotalWeight:  weight + coinbase.GetSize(),
		MerkleRoot:   merkleRoot,
		Version:      1,
	}

	// 8. Validate
	err := a.validator.Validate(tmpl)
	if err != nil {
		return nil, fmt.Errorf("assembly failed during validation: %w", err)
	}

	return tmpl, nil
}
