package mining

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// HashPipeline abstracts the hashing algorithm
type HashPipeline interface {
	HashHeader(header BlockHeader) string
}

type SHA256Pipeline struct{}

func NewSHA256Pipeline() *SHA256Pipeline {
	return &SHA256Pipeline{}
}

func (p *SHA256Pipeline) HashHeader(header BlockHeader) string {
	raw := fmt.Sprintf("%d:%s:%s:%d:%s:%d:%d",
		header.Version,
		header.PrevHash,
		header.MerkleRoot,
		header.Timestamp,
		header.Target,
		header.Nonce,
		header.ExtraNonce,
	)
	hashBytes := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(hashBytes[:])
}
