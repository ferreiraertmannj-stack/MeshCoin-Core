package mining

import "time"

type MiningPolicy struct {
	MaxWorkers    int
	QueueSize     int
	JobTimeout    time.Duration
	CacheTTL      time.Duration
	RefreshPeriod time.Duration
	MaxRetries    int
}

func DefaultMiningPolicy() MiningPolicy {
	return MiningPolicy{
		MaxWorkers:    4,
		QueueSize:     100,
		JobTimeout:    10 * time.Minute,
		CacheTTL:      60 * time.Second,
		RefreshPeriod: 15 * time.Second,
		MaxRetries:    3,
	}
}
