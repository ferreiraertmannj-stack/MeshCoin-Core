package sync

import (
	"errors"
	"sync"
	"time"
)

var (
	ErrSyncAlreadyRunning = errors.New("sync already running")
	ErrSyncNotRunning     = errors.New("sync not running")
)

// SyncManager coordinates the Fast Sync state machine, peer selection, and download progress.
type SyncManager struct {
	mu           sync.RWMutex
	state        SyncState
	localHeight  uint64
	remoteHeight uint64
	startTime    time.Time
	blocksSynced uint64
	isPaused     bool
	isCancelled  bool

	peerPool PeerPool
}

// NewSyncManager initializes the SyncManager architecture.
func NewSyncManager(pool PeerPool) *SyncManager {
	return &SyncManager{
		state:    StateIdle,
		peerPool: pool,
	}
}

// StartSync transitions the manager from Idle to discovering peers and starts the tracking.
func (sm *SyncManager) StartSync(targetHeight uint64) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.state != StateIdle && sm.state != StateFailed && sm.state != StateCompleted {
		return ErrSyncAlreadyRunning
	}

	sm.state = StateDiscoveringPeers
	sm.remoteHeight = targetHeight
	sm.startTime = time.Now()
	sm.blocksSynced = 0
	sm.isPaused = false
	sm.isCancelled = false

	return nil
}

// Pause pauses the current sync process.
func (sm *SyncManager) Pause() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	if sm.state == StateIdle || sm.state == StateCompleted || sm.state == StateFailed {
		return ErrSyncNotRunning
	}
	sm.isPaused = true
	return nil
}

// Resume unpauses the sync process.
func (sm *SyncManager) Resume() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	if !sm.isPaused {
		return nil
	}
	sm.isPaused = false
	return nil
}

// Cancel totally aborts the sync and resets to Idle.
func (sm *SyncManager) Cancel() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.isCancelled = true
	sm.state = StateIdle
	return nil
}

// Status returns a detailed report of the current sync state, speed, and ETA.
func (sm *SyncManager) Status() SyncStatusReport {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	pct := 0.0
	if sm.remoteHeight > 0 && sm.localHeight < sm.remoteHeight {
		pct = float64(sm.localHeight) / float64(sm.remoteHeight) * 100.0
	} else if sm.localHeight >= sm.remoteHeight && sm.remoteHeight > 0 {
		pct = 100.0
	}

	speed := 0.0
	eta := 0.0
	if sm.blocksSynced > 0 {
		elapsed := time.Since(sm.startTime).Seconds()
		if elapsed > 0 {
			speed = float64(sm.blocksSynced) / elapsed
			remaining := float64(sm.remoteHeight - sm.localHeight)
			if speed > 0 {
				eta = remaining / speed
			}
		}
	}

	return SyncStatusReport{
		State:          sm.state,
		LocalHeight:    sm.localHeight,
		RemoteHeight:   sm.remoteHeight,
		ProgressPct:    pct,
		ETASeconds:     eta,
		SpeedBlocksSec: speed,
	}
}

// UpdateLocalHeight increments the synced amount and updates the height.
func (sm *SyncManager) UpdateLocalHeight(height uint64) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	if height > sm.localHeight {
		sm.blocksSynced += (height - sm.localHeight)
		sm.localHeight = height
	}
}

// SetState modifies the internal state of the state machine.
func (sm *SyncManager) SetState(state SyncState) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.state = state
}
