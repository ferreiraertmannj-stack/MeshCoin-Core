package script

import "bytes"

type SignatureEngine interface {
	Verify(txHash []byte, pubKey interface{}, signature []byte) bool
	Sign(txHash []byte, privKey interface{}) ([]byte, error)
}

// DefaultSignatureEngine mock for Phase 47.
// It will replace the MockSignatureValidator from Phase 46.
type DefaultSignatureEngine struct{}

func NewDefaultSignatureEngine() *DefaultSignatureEngine {
	return &DefaultSignatureEngine{}
}

func (e *DefaultSignatureEngine) Verify(txHash []byte, pubKey interface{}, signature []byte) bool {
	// In a real implementation: ecdsa.Verify(pubKey.(*ecdsa.PublicKey), txHash, r, s)
	// For testing, we'll just check if signature matches pubkey bytes
	pubStr, ok := pubKey.(string)
	if !ok {
		return false
	}
	return bytes.Equal([]byte(pubStr), signature)
}

func (e *DefaultSignatureEngine) Sign(txHash []byte, privKey interface{}) ([]byte, error) {
	return []byte("mock_signature"), nil
}
