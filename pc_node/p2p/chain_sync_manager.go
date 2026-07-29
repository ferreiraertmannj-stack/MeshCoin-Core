package p2p

import (
	"context"
	"encoding/json"
	"sync"
	"time"
)

type BlockchainSyncManager struct {
	router       *MessageRouter
	peerManager  PeerManager
	inventoryMgr *InventoryManager
	provider     ChainProvider

	locator       *ChainLocator
	forkDetector  *ForkDetector
	scheduler     *BlockRequestScheduler
	pendingBlocks *PendingBlocksManager
	stats         *BlockchainSyncStatistics
	events        BlockchainSyncEvents

	mu     sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc
}

func NewBlockchainSyncManager(
	router *MessageRouter,
	pm PeerManager,
	invMgr *InventoryManager,
	provider ChainProvider,
	events BlockchainSyncEvents,
	reqTimeout time.Duration,
	maxRetries int,
) *BlockchainSyncManager {
	ctx, cancel := context.WithCancel(context.Background())
	stats := &BlockchainSyncStatistics{}

	return &BlockchainSyncManager{
		router:        router,
		peerManager:   pm,
		inventoryMgr:  invMgr,
		provider:      provider,
		locator:       NewChainLocator(provider),
		forkDetector:  NewForkDetector(provider, events),
		scheduler:     NewBlockRequestScheduler(router, pm, stats, events, reqTimeout, maxRetries),
		pendingBlocks: NewPendingBlocksManager(),
		stats:         stats,
		events:        events,
		ctx:           ctx,
		cancel:        cancel,
	}
}

func (sm *BlockchainSyncManager) Start() {
	sm.scheduler.Start()

	sm.router.RegisterHandler(MsgTypeGetHeaders, sm.handleGetHeaders)
	sm.router.RegisterHandler(MsgTypeHeaders, sm.handleHeaders)
	sm.router.RegisterHandler(MsgTypeGetBlocks, sm.handleGetBlocks)
}

func (sm *BlockchainSyncManager) Stop() {
	sm.cancel()
	sm.scheduler.Stop()

	sm.router.RemoveHandler(MsgTypeGetHeaders)
	sm.router.RemoveHandler(MsgTypeHeaders)
	sm.router.RemoveHandler(MsgTypeGetBlocks)
}

// RequestSync initiates a header-first sync with a specific peer
func (sm *BlockchainSyncManager) RequestSync(peer *Peer) {
	locators := sm.locator.BuildLocatorHashes()

	msg := P2PMessage{
		Type: MsgTypeGetHeaders,
		Payload: MsgGetHeaders{
			LocatorHashes: locators,
			HashStop:      "",
		},
	}
	_ = sm.router.SendToPeer(peer, msg)
}

func (sm *BlockchainSyncManager) handleGetHeaders(peer *Peer, msg P2PMessage) error {
	// Extract payload
	var req MsgGetHeaders
	data, _ := json.Marshal(msg.Payload)
	json.Unmarshal(data, &req)

	// In a real implementation, we would query the Ledger to find the common block
	// and return up to 2000 subsequent headers.
	// We simulate sending an empty list if not found.
	commonHash, _ := sm.locator.FindCommonBlock(req.LocatorHashes)

	_ = commonHash // Used to start fetching headers from Ledger

	resp := P2PMessage{
		Type: MsgTypeHeaders,
		Payload: MsgHeaders{
			Headers: []BlockHeader{}, // Provided by Ledger
		},
	}
	return sm.router.SendToPeer(peer, resp)
}

func (sm *BlockchainSyncManager) handleHeaders(peer *Peer, msg P2PMessage) error {
	var headersMsg MsgHeaders
	data, _ := json.Marshal(msg.Payload)
	json.Unmarshal(data, &headersMsg)

	if len(headersMsg.Headers) == 0 {
		return nil
	}

	sm.stats.IncHeadersReceived(len(headersMsg.Headers))
	if sm.events.OnHeadersReceived != nil {
		go sm.events.OnHeadersReceived(peer.Info().NodeID, len(headersMsg.Headers))
	}

	// Validate sequence
	validCount := 0
	for i := 0; i < len(headersMsg.Headers); i++ {
		header := headersMsg.Headers[i]

		// Check fork
		sm.forkDetector.CheckForFork(header.Hash, header.ParentHash, header.Height)

		// Determine if we need to request this block
		if !sm.provider.HasBlock(header.Hash) && !sm.pendingBlocks.HasBlock(header.Hash) {
			sm.scheduler.Schedule(header.Hash, header.Height)
		}

		validCount++
	}

	sm.stats.IncHeadersValidated(validCount)
	if sm.events.OnHeadersValidated != nil {
		go sm.events.OnHeadersValidated(peer.Info().NodeID, validCount)
	}

	return nil
}

func (sm *BlockchainSyncManager) handleGetBlocks(peer *Peer, msg P2PMessage) error {
	var req MsgGetBlocks
	data, _ := json.Marshal(msg.Payload)
	json.Unmarshal(data, &req)

	// Normally answers with MsgData containing blocks. Handled by InventoryManager's GetData flow.
	return nil
}

// HandleBlockReceived is called when Gossip or Inventory receives a block
func (sm *BlockchainSyncManager) HandleBlockReceived(hash string, parentHash string, height uint64, rawData []byte) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.stats.IncBlocksReceived(1)
	if sm.events.OnBlockReceived != nil {
		go sm.events.OnBlockReceived("unknown", hash)
	}

	sm.scheduler.MarkReceived(hash)

	// Check if we have the parent
	if !sm.provider.HasBlock(parentHash) {
		// Orphan block
		sm.pendingBlocks.AddBlock(PendingBlock{
			Hash:       hash,
			ParentHash: parentHash,
			Height:     height,
			Data:       rawData,
		})
		sm.stats.SetOrphanBlocks(sm.pendingBlocks.Count())

		// Schedule parent fetch
		sm.scheduler.Schedule(parentHash, height-1)
		return
	}

	// We have parent, we can import
	sm.importBlock(hash, parentHash, height, rawData)
}

func (sm *BlockchainSyncManager) importBlock(hash string, parentHash string, height uint64, rawData []byte) {
	// Call OnBlockImported (mocking the connection to real importer)
	sm.stats.IncBlocksImported(1)
	if sm.events.OnBlockImported != nil {
		sm.events.OnBlockImported(hash, height)
	}

	// Process children waiting for this block
	children := sm.pendingBlocks.GetChildren(hash)
	sm.pendingBlocks.RemoveBlock(hash)
	sm.stats.SetOrphanBlocks(sm.pendingBlocks.Count())

	for _, child := range children {
		sm.importBlock(child.Hash, child.ParentHash, child.Height, child.Data)
	}
}

func (sm *BlockchainSyncManager) Statistics() BlockchainSyncStatistics {
	return sm.stats.Snapshot()
}
