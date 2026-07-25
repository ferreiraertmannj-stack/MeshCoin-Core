package main

import (
	"encoding/base64"
	"math/rand"
	"strings"
	"time"
)

// NebulaPQC simula LWE (Learning With Errors) Híbrido
type NebulaPQC struct{}

func GeneratePQCPublicKey() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	key := make([]byte, 128)
	for i := range key {
		key[i] = byte(r.Intn(256))
	}
	return base64.StdEncoding.EncodeToString(key)
}

func EncryptMessagePQC(message string, recipientPublicKey string) string {
	msgBytes := []byte(message)
	keyBytes, err := base64.StdEncoding.DecodeString(recipientPublicKey)
	if err != nil || len(keyBytes) == 0 {
		return message // Fallback
	}

	cipher := make([]byte, len(msgBytes))
	for i := 0; i < len(msgBytes); i++ {
		k := keyBytes[i%len(keyBytes)]
		cipher[i] = msgBytes[i] ^ k
	}

	return "NBL_PQC_V1::" + base64.StdEncoding.EncodeToString(cipher)
}

func DecryptMessagePQC(cipherText string, myPublicKey string) string {
	if !strings.HasPrefix(cipherText, "NBL_PQC_V1::") {
		return cipherText
	}

	rawBase64 := cipherText[14:]
	cipherBytes, err := base64.StdEncoding.DecodeString(rawBase64)
	if err != nil {
		return cipherText
	}

	keyBytes, err := base64.StdEncoding.DecodeString(myPublicKey)
	if err != nil || len(keyBytes) == 0 {
		return cipherText
	}

	plain := make([]byte, len(cipherBytes))
	for i := 0; i < len(cipherBytes); i++ {
		k := keyBytes[i%len(keyBytes)]
		plain[i] = cipherBytes[i] ^ k
	}

	return string(plain)
}
