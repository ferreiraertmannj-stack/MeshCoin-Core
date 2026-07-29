package sync

// P2P messages specific to Fast Sync for Nebula Network.
// These are completely decoupled from the existing block-by-block protocol.

// MsgGetHeaders requests a list of header metadata starting from a specific hash.
type MsgGetHeaders struct {
	StartHash string
	Limit     int
}

// MsgHeaders is the response containing the requested headers.
type MsgHeaders struct {
	Headers []HeaderMetadata
}

// HeaderMetadata contains a lightweight representation of a block for quick verification.
type HeaderMetadata struct {
	Index uint64
	Hash  string
}

// MsgGetBlocks requests a range of blocks by their indices.
type MsgGetBlocks struct {
	StartIndex uint64
	EndIndex   uint64
}

// MsgBlocks is the response containing the raw byte data of the requested blocks.
type MsgBlocks struct {
	Blocks [][]byte // Raw block data directly readable by the Storage Engine
}

// MsgSyncStatus is used to share and query the sync status with peers.
type MsgSyncStatus struct {
	CurrentHeight uint64
	IsSyncing     bool
}

// MsgSyncCompleted is broadcasted to peers when the local Fast Sync finishes.
type MsgSyncCompleted struct {
	NodeID string
}
