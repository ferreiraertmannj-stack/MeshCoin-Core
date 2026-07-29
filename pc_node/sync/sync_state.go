package sync

import "time"

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
	CurrentState     SyncState
	CurrentHeight    uint64
	RemoteHeight     uint64
	DownloadedBlocks uint64
	DownloadedChunks int
	PendingChunks    int
	FailedChunks     int
	Workers          int
	Peers            int
	SpeedBlocksSec   float64
	ETASeconds       float64
	ProgressPercent  float64
	LastError        string
	UpdatedAt        time.Time
}
