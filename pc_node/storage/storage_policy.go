package storage

import "time"

type StoragePolicy struct {
	FlushInterval    time.Duration
	CacheSize        int
	CacheTTL         time.Duration
	QueueSize        int
	SnapshotInterval time.Duration
	Compression      bool
	SyncWrites       bool
}

func DefaultStoragePolicy() StoragePolicy {
	return StoragePolicy{
		FlushInterval:    5 * time.Second,
		CacheSize:        10000,
		CacheTTL:         60 * time.Second,
		QueueSize:        5000,
		SnapshotInterval: 24 * time.Hour,
		Compression:      true,
		SyncWrites:       false,
	}
}
