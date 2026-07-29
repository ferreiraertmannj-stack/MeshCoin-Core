package mempool

import (
	"sync"
	"time"
)

type entry struct {
	tx      Transaction
	addedAt time.Time
}

type MempoolCache struct {
	// Indices
	txs      map[string]entry           // Hash -> Transaction
	bySender map[string]map[string]bool // Sender -> Hash set

	// Ordered access structures could be added here if needed,
	// but standard Maps provide O(1) for our specific required lookups

	mu sync.RWMutex
}

func NewMempoolCache() *MempoolCache {
	return &MempoolCache{
		txs:      make(map[string]entry),
		bySender: make(map[string]map[string]bool),
	}
}

func (c *MempoolCache) Add(tx Transaction) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	hash := tx.GetHash()
	if _, exists := c.txs[hash]; exists {
		return false
	}

	c.txs[hash] = entry{
		tx:      tx,
		addedAt: time.Now(),
	}

	sender := tx.GetSender()
	if c.bySender[sender] == nil {
		c.bySender[sender] = make(map[string]bool)
	}
	c.bySender[sender][hash] = true

	return true
}

func (c *MempoolCache) Get(hash string) (Transaction, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	e, ok := c.txs[hash]
	if !ok {
		return nil, false
	}
	return e.tx, true
}

func (c *MempoolCache) GetAddedAt(hash string) time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if e, ok := c.txs[hash]; ok {
		return e.addedAt
	}
	return time.Time{}
}

func (c *MempoolCache) Contains(hash string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	_, ok := c.txs[hash]
	return ok
}

func (c *MempoolCache) Remove(hash string) (Transaction, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	e, ok := c.txs[hash]
	if !ok {
		return nil, false
	}

	delete(c.txs, hash)

	sender := e.tx.GetSender()
	if set, exists := c.bySender[sender]; exists {
		delete(set, hash)
		if len(set) == 0 {
			delete(c.bySender, sender)
		}
	}

	return e.tx, true
}

func (c *MempoolCache) Count() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.txs)
}

func (c *MempoolCache) GetBySender(sender string) []Transaction {
	c.mu.RLock()
	defer c.mu.RUnlock()

	set, ok := c.bySender[sender]
	if !ok {
		return nil
	}

	res := make([]Transaction, 0, len(set))
	for hash := range set {
		if e, exists := c.txs[hash]; exists {
			res = append(res, e.tx)
		}
	}
	return res
}

func (c *MempoolCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.txs = make(map[string]entry)
	c.bySender = make(map[string]map[string]bool)
}

func (c *MempoolCache) GetAll() []Transaction {
	c.mu.RLock()
	defer c.mu.RUnlock()

	res := make([]Transaction, 0, len(c.txs))
	for _, e := range c.txs {
		res = append(res, e.tx)
	}
	return res
}
