package sync

// SyncState represents the various phases of the Fast Sync state machine.
type SyncState string

const (
	StateIdle              SyncState = "Idle"
	StateDiscoveringPeers  SyncState = "DiscoveringPeers"
	StateRequestingHeaders SyncState = "RequestingHeaders"
	StateDownloadingBlocks SyncState = "DownloadingBlocks"
	StateVerifyingBlocks   SyncState = "VerifyingBlocks"
	StateImportingBlocks   SyncState = "ImportingBlocks"
	StateCompleted         SyncState = "Completed"
	StateFailed            SyncState = "Failed"
)

// SyncStatusReport encapsulates the current progress of the fast sync operation.
type SyncStatusReport struct {
	State          SyncState
	LocalHeight    uint64
	RemoteHeight   uint64
	ProgressPct    float64
	ETASeconds     float64
	SpeedBlocksSec float64
}
