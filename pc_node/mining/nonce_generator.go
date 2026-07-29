package mining

import (
	"math/rand"
	"sync"
	"time"
)

type NonceGenerator struct {
	mu  sync.Mutex
	rng *rand.Rand
}

func NewNonceGenerator() *NonceGenerator {
	return &NonceGenerator{
		rng: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (g *NonceGenerator) Generate() (uint64, uint64) {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.rng.Uint64(), g.rng.Uint64()
}
