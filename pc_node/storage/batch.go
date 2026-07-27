package storage

// Batch defines atomic transactions for the storage engine.
// All operations within a batch must succeed or fail together.
type Batch interface {
	// PutBlock stores a serialized block at a specific index.
	PutBlock(index uint64, blockData []byte) error

	// PutBalance updates the UTXO / Balance index for an address.
	PutBalance(address string, balance float64) error

	// Commit applies all the operations atomically.
	Commit() error

	// Discard drops all pending operations in this batch.
	Discard()
}
