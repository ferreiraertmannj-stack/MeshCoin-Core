package p2p

import (
	"sync"
)

// PendingBlock represents a block waiting for its parent to be imported
type PendingBlock struct {
	Hash       string
	ParentHash string
	Height     uint64
	Data       []byte
}

type PendingBlocksManager struct {
	// parentHash -> list of blocks waiting for it
	waitingForParent map[string][]PendingBlock
	// hash -> block (fast lookup)
	orphans map[string]PendingBlock

	mu sync.RWMutex
}

func NewPendingBlocksManager() *PendingBlocksManager {
	return &PendingBlocksManager{
		waitingForParent: make(map[string][]PendingBlock),
		orphans:          make(map[string]PendingBlock),
	}
}

func (pm *PendingBlocksManager) AddBlock(block PendingBlock) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// If already have it, do nothing
	if _, exists := pm.orphans[block.Hash]; exists {
		return
	}

	pm.orphans[block.Hash] = block
	pm.waitingForParent[block.ParentHash] = append(pm.waitingForParent[block.ParentHash], block)
}

func (pm *PendingBlocksManager) HasBlock(hash string) bool {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	_, exists := pm.orphans[hash]
	return exists
}

// GetChildren returns any blocks that were waiting for the given parent hash
func (pm *PendingBlocksManager) GetChildren(parentHash string) []PendingBlock {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	children, exists := pm.waitingForParent[parentHash]
	if !exists {
		return nil
	}

	// We do not remove them here. We remove them once they are successfully processed via RemoveBlock.
	// But actually, if they are returned, the caller will attempt to import them.
	// We can just return a copy.
	cp := make([]PendingBlock, len(children))
	copy(cp, children)
	return cp
}

func (pm *PendingBlocksManager) RemoveBlock(hash string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	block, exists := pm.orphans[hash]
	if !exists {
		return
	}

	delete(pm.orphans, hash)

	// Remove from waitingForParent
	children := pm.waitingForParent[block.ParentHash]
	for i, c := range children {
		if c.Hash == hash {
			// Remove element
			pm.waitingForParent[block.ParentHash] = append(children[:i], children[i+1:]...)
			break
		}
	}

	// Clean up empty slices
	if len(pm.waitingForParent[block.ParentHash]) == 0 {
		delete(pm.waitingForParent, block.ParentHash)
	}
}

func (pm *PendingBlocksManager) Count() int {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return len(pm.orphans)
}
