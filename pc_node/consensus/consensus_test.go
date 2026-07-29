package consensus

import (
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type mockTx struct {
	hash   string
	reward uint64
	fees   uint64
	size   uint64
}

func (m *mockTx) GetHash() string   { return m.hash }
func (m *mockTx) GetReward() uint64 { return m.reward }
func (m *mockTx) GetFees() uint64   { return m.fees }
func (m *mockTx) GetSize() uint64   { return m.size }

type mockDifficulty struct{}

func (m *mockDifficulty) UpdateDifficulty(currentHeight uint64) (string, string, bool) {
	return "0000000000000000000000000000000000000000000000000000000000000000", "0", false
}
func (m *mockDifficulty) GetCurrentTarget() string {
	return "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"
}
func (m *mockDifficulty) GetCurrentDifficulty() string { return "1" }

type mockNetwork struct{}

func (m *mockNetwork) GetNetworkTimestamp() int64 { return time.Now().Unix() }

type mockAppender struct {
	height uint64
	mu     sync.Mutex
}

func (m *mockAppender) AppendBlock(block *Block) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.height = block.Height
	return nil
}
func (m *mockAppender) GetHighestBlockHeight() uint64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.height
}

func calculateTestMerkle(coinbase string, txs []Transaction) string {
	raw := coinbase
	for _, tx := range txs {
		raw += tx.GetHash()
	}
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}

func createValidBlock(policy ConsensusPolicy) *Block {
	txs := []Transaction{
		&mockTx{hash: "tx1", fees: 10, size: 100},
	}

	// Genesis halving = 0 -> subsidy = 50 * 10^8
	coinbase := &mockTx{
		hash:   "cb",
		reward: policy.InitialSubsidy,
		fees:   10,
		size:   100,
	}

	root := calculateTestMerkle(coinbase.GetHash(), txs)

	return &Block{
		Version:      policy.BlockVersion,
		PrevHash:     "prev",
		MerkleRoot:   root,
		Timestamp:    time.Now().Unix(),
		Target:       "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
		Nonce:        0, // with this target, hash will always pass
		ExtraNonce:   0,
		Coinbase:     coinbase,
		Transactions: txs,
		Height:       1,
	}
}

func TestBlockAcceptance_ValidBlock(t *testing.T) {
	policy := DefaultConsensusPolicy()
	events := ConsensusEvents{}
	stats := &ConsensusStatistics{}

	rewardVal := NewRewardValidator(policy)
	cbVal := NewCoinbaseValidator(rewardVal)
	diff := &mockDifficulty{}

	validator := NewBlockValidator(policy, cbVal, diff)
	appender := &mockAppender{}
	chainUp := NewChainUpdater(appender, events)

	pipeline := NewBlockAcceptancePipeline(validator, chainUp, diff, stats, events, &mockNetwork{}, appender)

	block := createValidBlock(policy)

	err := pipeline.ProcessBlock(block)
	if err != nil {
		t.Fatalf("Expected valid block, got error: %v", err)
	}

	if stats.BlocksAccepted != 1 {
		t.Fatalf("Expected 1 accepted")
	}
}

func TestBlockAcceptance_InvalidMerkle(t *testing.T) {
	policy := DefaultConsensusPolicy()
	stats := &ConsensusStatistics{}

	rewardVal := NewRewardValidator(policy)
	cbVal := NewCoinbaseValidator(rewardVal)
	diff := &mockDifficulty{}

	validator := NewBlockValidator(policy, cbVal, diff)
	appender := &mockAppender{}
	chainUp := NewChainUpdater(appender, ConsensusEvents{})

	pipeline := NewBlockAcceptancePipeline(validator, chainUp, diff, stats, ConsensusEvents{}, &mockNetwork{}, appender)

	block := createValidBlock(policy)
	block.MerkleRoot = "deadbeef" // Tampered

	err := pipeline.ProcessBlock(block)
	if err == nil {
		t.Fatalf("Expected invalid merkle error")
	}
	if stats.BlocksRejected != 1 {
		t.Fatalf("Expected 1 rejected")
	}
}

func TestBlockAcceptance_InvalidReward(t *testing.T) {
	policy := DefaultConsensusPolicy()
	stats := &ConsensusStatistics{}

	rewardVal := NewRewardValidator(policy)
	cbVal := NewCoinbaseValidator(rewardVal)
	diff := &mockDifficulty{}

	validator := NewBlockValidator(policy, cbVal, diff)
	appender := &mockAppender{}
	chainUp := NewChainUpdater(appender, ConsensusEvents{})

	pipeline := NewBlockAcceptancePipeline(validator, chainUp, diff, stats, ConsensusEvents{}, &mockNetwork{}, appender)

	block := createValidBlock(policy)

	// Tamper reward
	block.Coinbase = &mockTx{
		hash:   "cb",
		reward: policy.InitialSubsidy + 1, // Invalid
		fees:   10,
		size:   100,
	}
	// Update merkle root for new cb
	block.MerkleRoot = calculateTestMerkle(block.Coinbase.GetHash(), block.Transactions)

	err := pipeline.ProcessBlock(block)
	if err == nil {
		t.Fatalf("Expected invalid reward error")
	}
}

func TestBlockAcceptance_ExceededWeight(t *testing.T) {
	policy := DefaultConsensusPolicy()
	policy.MaxBlockWeight = 100 // Very small

	stats := &ConsensusStatistics{}

	rewardVal := NewRewardValidator(policy)
	cbVal := NewCoinbaseValidator(rewardVal)
	diff := &mockDifficulty{}

	validator := NewBlockValidator(policy, cbVal, diff)
	appender := &mockAppender{}
	chainUp := NewChainUpdater(appender, ConsensusEvents{})

	pipeline := NewBlockAcceptancePipeline(validator, chainUp, diff, stats, ConsensusEvents{}, &mockNetwork{}, appender)

	block := createValidBlock(policy)

	err := pipeline.ProcessBlock(block)
	if err == nil {
		t.Fatalf("Expected weight exceeded error")
	}
}

func TestBlockAcceptance_StressTest(t *testing.T) {
	policy := DefaultConsensusPolicy()
	stats := &ConsensusStatistics{}

	rewardVal := NewRewardValidator(policy)
	cbVal := NewCoinbaseValidator(rewardVal)
	diff := &mockDifficulty{}

	validator := NewBlockValidator(policy, cbVal, diff)
	appender := &mockAppender{}

	var updated atomic.Int32
	events := ConsensusEvents{
		OnChainUpdated: func(height uint64) {
			updated.Add(1)
		},
	}

	chainUp := NewChainUpdater(appender, events)
	pipeline := NewBlockAcceptancePipeline(validator, chainUp, diff, stats, events, &mockNetwork{}, appender)

	var wg sync.WaitGroup
	routines := 1000

	// Pre-build 1000 valid blocks
	blocks := make([]*Block, routines)
	for i := 0; i < routines; i++ {
		b := createValidBlock(policy)
		b.Height = uint64(i + 1)
		blocks[i] = b
	}

	for i := 0; i < routines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			_ = pipeline.ProcessBlock(blocks[idx])
		}(i)
	}

	wg.Wait()

	// Allow callbacks to finish
	time.Sleep(100 * time.Millisecond)

	if stats.BlocksAccepted != 1000 {
		t.Fatalf("Expected 1000 accepted blocks, got %d", stats.BlocksAccepted)
	}

	if updated.Load() != 1000 {
		t.Fatalf("Expected 1000 chain update events, got %d", updated.Load())
	}
}
