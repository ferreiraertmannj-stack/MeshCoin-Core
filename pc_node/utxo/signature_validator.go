package utxo

// SignatureValidator is an interface to allow injecting ECDSA, Ed25519, etc.
type SignatureValidator interface {
	Verify(txHash string, script []byte, signature []byte) bool
}

// MockSignatureValidator for Phase 46
type MockSignatureValidator struct{}

func (m *MockSignatureValidator) Verify(txHash string, script []byte, signature []byte) bool {
	return true
}
