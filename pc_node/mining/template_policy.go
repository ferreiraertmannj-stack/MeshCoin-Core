package mining

import "time"

type TemplatePolicy struct {
	MaxBlockWeight  uint64
	MaxBlockSize    uint64
	MaxSigOps       uint64
	MaxTransactions int
	BlockReward     uint64
	CacheTTL        time.Duration
	RefreshInterval time.Duration
}

func DefaultTemplatePolicy() TemplatePolicy {
	return TemplatePolicy{
		MaxBlockWeight:  4000000,
		MaxBlockSize:    1000000,
		MaxSigOps:       20000,
		MaxTransactions: 5000,
		BlockReward:     50 * 100000000, // 50 Coins
		CacheTTL:        30 * time.Second,
		RefreshInterval: 5 * time.Second,
	}
}
