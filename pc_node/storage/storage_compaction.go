package storage

import (
	"context"
	"time"
)

type StorageCompactor struct {
	stats *StorageStatistics
}

func NewStorageCompactor(stats *StorageStatistics) *StorageCompactor {
	return &StorageCompactor{
		stats: stats,
	}
}

// Compact triggers a background compaction process.
// In a real database (like LevelDB/RocksDB) this merges SSTables.
func (c *StorageCompactor) Compact(ctx context.Context) error {
	// Simulate compaction work
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(10 * time.Millisecond):
		c.stats.IncCompactions()
	}
	return nil
}
