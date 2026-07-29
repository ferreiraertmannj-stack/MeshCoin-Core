package script

import (
	"crypto/sha256"
)

type HashEngine interface {
	Hash256(data []byte) []byte
	Hash160(data []byte) []byte
}

type DefaultHashEngine struct{}

func NewDefaultHashEngine() *DefaultHashEngine {
	return &DefaultHashEngine{}
}

// Hash256 is double SHA-256
func (e *DefaultHashEngine) Hash256(data []byte) []byte {
	h1 := sha256.Sum256(data)
	h2 := sha256.Sum256(h1[:])
	return h2[:]
}

// Hash160 mock using SHA256 truncated to 20 bytes
func (e *DefaultHashEngine) Hash160(data []byte) []byte {
	h1 := sha256.Sum256(data)
	return h1[:20] // Mock for RIPEMD160 length
}
