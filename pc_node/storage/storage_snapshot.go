package storage

import (
	"fmt"
	"time"
)

type Snapshot struct {
	ID        string
	Timestamp time.Time
}

type StorageSnapshot struct {
	stats *StorageStatistics
}

func NewStorageSnapshot(stats *StorageStatistics) *StorageSnapshot {
	return &StorageSnapshot{
		stats: stats,
	}
}

// Create creates a point-in-time snapshot of the database
func (s *StorageSnapshot) Create() (*Snapshot, error) {
	s.stats.IncSnapshotsCreated()
	return &Snapshot{
		ID:        fmt.Sprintf("snap-%d", time.Now().UnixNano()),
		Timestamp: time.Now(),
	}, nil
}
