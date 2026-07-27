package storage

// Engine defines the core contract for the persistence layer.
// It abstracts away the underlying database (JSON, BadgerDB, etc).
type Engine interface {
	// Lifecycle
	Open(connectionString string) error
	Close() error

	// Batch Operations
	NewBatch() Batch

	// Block Read Operations
	GetBlockByIndex(index uint64) ([]byte, error)
	GetLatestBlock() ([]byte, error)

	// State Read Operations
	GetBalance(address string) (float64, error)

	// Iteration
	NewBlockIterator() Iterator

	// Snapshots
	CreateSnapshot(path string) error
}
