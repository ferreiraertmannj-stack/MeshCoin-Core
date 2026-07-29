package script

import (
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"time"
)

type ScriptCache struct {
	cache map[string]time.Time
	ttl   time.Duration
	mu    sync.RWMutex
}

func NewScriptCache(ttl time.Duration) *ScriptCache {
	return &ScriptCache{
		cache: make(map[string]time.Time),
		ttl:   ttl,
	}
}

func (c *ScriptCache) makeKey(scriptSig, scriptPubKey, txHash []byte) string {
	h := sha256.New()
	h.Write(scriptSig)
	h.Write(scriptPubKey)
	h.Write(txHash)
	return hex.EncodeToString(h.Sum(nil))
}

func (c *ScriptCache) Get(scriptSig, scriptPubKey, txHash []byte) bool {
	key := c.makeKey(scriptSig, scriptPubKey, txHash)
	c.mu.RLock()
	defer c.mu.RUnlock()

	ts, ok := c.cache[key]
	if !ok {
		return false
	}
	if time.Since(ts) > c.ttl {
		return false
	}
	return true
}

func (c *ScriptCache) Set(scriptSig, scriptPubKey, txHash []byte) {
	key := c.makeKey(scriptSig, scriptPubKey, txHash)
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache[key] = time.Now()
}
