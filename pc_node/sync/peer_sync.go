package sync

// Peer defines the contract for interacting with remote nodes during Fast Sync.
// This interface allows future expansion to support multiple peers concurrently.
type Peer interface {
	ID() string
	SendMsg(msg interface{}) error
	RequestHeaders(startHash string, limit int) error
	RequestBlocks(startIndex, endIndex uint64) error
	Disconnect()
}

// PeerPool manages available peers for sync, supporting selection and load balancing.
type PeerPool interface {
	GetBestPeer() Peer
	GetAllPeers() []Peer
	AddPeer(p Peer)
	RemovePeer(id string)
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
