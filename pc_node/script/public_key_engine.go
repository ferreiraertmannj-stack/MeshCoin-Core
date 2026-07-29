package script

import "fmt"

type PublicKeyEngine interface {
	ParsePubKey(pubKeyBytes []byte) (interface{}, error)
}

type DefaultPublicKeyEngine struct{}

func NewDefaultPublicKeyEngine() *DefaultPublicKeyEngine {
	return &DefaultPublicKeyEngine{}
}

func (e *DefaultPublicKeyEngine) ParsePubKey(pubKeyBytes []byte) (interface{}, error) {
	// Mock implementation for ECDSA / Ed25519 public key parsing
	if len(pubKeyBytes) == 0 {
		return nil, fmt.Errorf("empty public key")
	}
	return string(pubKeyBytes), nil
}
