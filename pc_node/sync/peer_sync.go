package sync

import "time"

// Peer defines the contract for interacting with remote nodes during Fast Sync.
// This interface allows future expansion to support multiple peers concurrently.
type Peer interface {
	ID() string
	SendMsg(msg interface{}) error
	RequestHeaders(startHash string, limit int) error
	RequestBlocks(startIndex, endIndex uint64) error
	Disconnect()

	// Metrics for PeerPool scoring
	Height() uint64
	Latency() time.Duration
	Failures() int
	AddFailure()
	ConnectedSince() time.Time
}

// PeerPool gerencia os peers para sincronização com Thread Safety e algoritmos de Score.
type PeerPool interface {
	AddPeer(p Peer)
	RemovePeer(id string)
	PeerCount() int
	ListPeers() []Peer
	BestPeer() Peer
	RandomPeer() Peer
	FastestPeer() Peer
	HighestPeer() Peer
}

// BlockValidator defines the contract for validating blocks before importing them.
// This supports incremental validation during the VerifyingBlocks state.
type BlockValidator interface {
	ValidateHeader(header HeaderMetadata) error
	ValidateBlock(blockData []byte) error
}

// SnapshotManager exposes interfaces for future snapshot-based synchronization.
type SnapshotManager interface {
	ImportSnapshot(path string) error
	ExportSnapshot(path string) error
}
