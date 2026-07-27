package storage

import "errors"

var (
	// ErrNotFound is returned when a requested key (block, balance) does not exist in the storage.
	ErrNotFound = errors.New("storage: key not found")

	// ErrClosed is returned when an operation is attempted on a closed database or iterator.
	ErrClosed = errors.New("storage: database closed")

	// ErrInvalidKey is returned when the provided key or index is malformed.
	ErrInvalidKey = errors.New("storage: invalid key format")

	// ErrCorruptedData is returned when the retrieved data fails integrity checks.
	ErrCorruptedData = errors.New("storage: corrupted data encountered")

	// ErrUnsupported is returned when a specific storage backend does not support an operation.
	ErrUnsupported = errors.New("storage: operation not supported by this engine")
)
