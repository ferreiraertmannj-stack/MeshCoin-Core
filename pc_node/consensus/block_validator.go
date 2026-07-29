package consensus

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
)

// Interfaces for isolation
type Block struct {
	Version      uint32
	PrevHash     string
	MerkleRoot   string
	Timestamp    int64
	Target       string
	Nonce        uint64
	ExtraNonce   uint64
	Coinbase     Transaction
	Transactions []Transaction
	Height       uint64
}

func (b *Block) Hash() string {
	raw := fmt.Sprintf("%d:%s:%s:%d:%s:%d:%d",
		b.Version, b.PrevHash, b.MerkleRoot, b.Timestamp, b.Target, b.Nonce, b.ExtraNonce,
	)
	hashBytes := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(hashBytes[:])
}

type Transaction interface {
	GetHash() string
	GetReward() uint64
	GetFees() uint64
	GetSize() uint64
}

type BlockValidator struct {
	policy            ConsensusPolicy
	coinbaseValidator *CoinbaseValidator
	difficulty        DifficultyUpdater
}

func NewBlockValidator(policy ConsensusPolicy, cbValidator *CoinbaseValidator, diff DifficultyUpdater) *BlockValidator {
	return &BlockValidator{
		policy:            policy,
		coinbaseValidator: cbValidator,
		difficulty:        diff,
	}
}

func (v *BlockValidator) Validate(block *Block, currentNetworkTime int64, prevHash string) error {
	// 1. Version
	if block.Version != v.policy.BlockVersion {
		return fmt.Errorf("invalid block version")
	}

	// 2. PrevHash
	if block.Height > 0 && block.PrevHash != prevHash {
		return fmt.Errorf("prevHash mismatch")
	}

	// 3. Timestamp
	if block.Timestamp > currentNetworkTime+int64(v.policy.MaxFutureTime.Seconds()) {
		return fmt.Errorf("timestamp too far in the future")
	}

	// 4. Target & Difficulty
	if block.Target != v.difficulty.GetCurrentTarget() {
		return fmt.Errorf("invalid target")
	}

	// 5. PoW Hash
	hash := block.Hash()
	hashInt := new(big.Int)
	hashInt.SetString(hash, 16)
	targetInt := new(big.Int)
	targetInt.SetString(block.Target, 16)

	if hashInt.Cmp(targetInt) > 0 {
		return fmt.Errorf("hash does not meet target")
	}

	// 6. Weight
	totalWeight := block.Coinbase.GetSize()
	for _, tx := range block.Transactions {
		totalWeight += tx.GetSize()
	}
	if totalWeight > v.policy.MaxBlockWeight {
		return fmt.Errorf("block exceeds max weight")
	}

	// 7. Merkle Root
	hashes := []string{block.Coinbase.GetHash()}
	for _, tx := range block.Transactions {
		hashes = append(hashes, tx.GetHash())
	}
	expectedRoot := calculateMerkleRoot(hashes)
	if block.MerkleRoot != expectedRoot {
		return fmt.Errorf("invalid merkle root")
	}

	// 8. Coinbase & Duplicates
	dedup := make(map[string]bool)
	totalFees := uint64(0)

	for _, tx := range block.Transactions {
		h := tx.GetHash()
		if dedup[h] {
			return fmt.Errorf("duplicate transaction %s", h)
		}
		dedup[h] = true
		totalFees += tx.GetFees()
	}

	err := v.coinbaseValidator.Validate(block.Height, totalFees, block.Coinbase)
	if err != nil {
		return fmt.Errorf("coinbase validation failed: %w", err)
	}

	return nil
}

func calculateMerkleRoot(hashes []string) string {
	raw := ""
	for _, h := range hashes {
		raw += h
	}
	hashBytes := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(hashBytes[:])
}
