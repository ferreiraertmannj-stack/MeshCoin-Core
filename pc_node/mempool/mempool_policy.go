package mempool

import "time"

type MempoolPolicy struct {
	MaxTransactions int
	MaxMemoryBytes  uint64
	TTL             time.Duration
	AllowRBF        bool // Replace By Fee
	MinFee          uint64
}

func DefaultMempoolPolicy() MempoolPolicy {
	return MempoolPolicy{
		MaxTransactions: 50000,
		MaxMemoryBytes:  100 * 1024 * 1024, // 100 MB
		TTL:             24 * time.Hour,
		AllowRBF:        true,
		MinFee:          1,
	}
}
