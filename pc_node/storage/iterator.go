package storage

// Iterator provides sequential access to stored blocks without loading the entire chain into memory.
type Iterator interface {
	// Next advances the iterator to the next block. Returns false if no more blocks exist.
	Next() bool

	// Value returns the serialized bytes of the current block.
	Value() []byte

	// Error returns any error encountered during iteration.
	Error() error

	// Close releases resources associated with the iterator.
	Close()
}
