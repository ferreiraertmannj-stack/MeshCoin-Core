package mining

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

// CoinbaseTx implements the Transaction interface for the builder
type CoinbaseTx struct {
	Hash      string
	Fee       uint64
	Size      uint64
	Sender    string
	Timestamp int64
	Reward    uint64
	Fees      uint64
}

func (c *CoinbaseTx) GetHash() string     { return c.Hash }
func (c *CoinbaseTx) GetFee() uint64      { return c.Fee }
func (c *CoinbaseTx) GetSize() uint64     { return c.Size }
func (c *CoinbaseTx) GetSender() string   { return c.Sender }
func (c *CoinbaseTx) GetTimestamp() int64 { return c.Timestamp }

type CoinbaseBuilder struct {
	policy TemplatePolicy
}

func NewCoinbaseBuilder(policy TemplatePolicy) *CoinbaseBuilder {
	return &CoinbaseBuilder{
		policy: policy,
	}
}

func (b *CoinbaseBuilder) Build(height uint64, totalFees uint64) Transaction {
	// Simple deterministic hash for abstraction
	raw := fmt.Sprintf("COINBASE-%d-%d-%d", height, totalFees, time.Now().UnixNano())
	hashBytes := sha256.Sum256([]byte(raw))
	hashStr := hex.EncodeToString(hashBytes[:])

	return &CoinbaseTx{
		Hash:      hashStr,
		Fee:       0,   // Coinbase pays no fee
		Size:      100, // Arbitrary weight for abstraction
		Sender:    "SYSTEM",
		Timestamp: time.Now().Unix(),
		Reward:    b.policy.BlockReward,
		Fees:      totalFees,
	}
}
