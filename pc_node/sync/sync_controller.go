package sync

import (
	"context"
	"sync"
	"time"
)

// SyncControllerEventHandlers contains decoupled callbacks for observing Sync state.
type SyncControllerEventHandlers struct {
	OnStateChanged      func(oldState, newState SyncState)
	OnDownloadStarted   func()
	OnDownloadCompleted func()
	OnFailed            func(err error)
	OnCancelled         func()
}

// SyncController binds the SyncManager (State) to the Downloader (Action)
// and handles the full Fast Sync state machine lifecycle.
type SyncController struct {
	mu           sync.RWMutex
	manager      *SyncManager
	downloader   *Downloader
	events       SyncControllerEventHandlers
	
	cancelFunc   context.CancelFunc
	isActive     bool
}

// NewSyncController initializes the main coordinator for Fast Sync.
func NewSyncController(manager *SyncManager, downloader *Downloader, events SyncControllerEventHandlers) *SyncController {
	return &SyncController{
		manager:    manager,
		downloader: downloader,
		events:     events,
	}
}

// Start initiates the full sync loop.
func (c *SyncController) Start(targetHeight uint64) error {
	c.mu.Lock()
	if c.isActive {
		c.mu.Unlock()
		return ErrSyncAlreadyRunning
	}
	ctx, cancel := context.WithCancel(context.Background())
	c.cancelFunc = cancel
	c.isActive = true
	c.mu.Unlock()

	err := c.manager.StartSync(targetHeight)
	if err != nil {
		c.mu.Lock()
		c.isActive = false
		c.mu.Unlock()
		return err
	}

	go c.runStateMachine(ctx, targetHeight)
	return nil
}

// Pause pauses both the State Manager and the Downloader immediately.
func (c *SyncController) Pause() error {
	err := c.manager.Pause()
	if err == nil {
		c.downloader.Pause()
	}
	return err
}

// Resume restarts both State Manager and Downloader.
func (c *SyncController) Resume() error {
	err := c.manager.Resume()
	if err == nil {
		c.downloader.Resume()
	}
	return err
}

// Stop cleanly aborts the sync without emitting a failure (acting like Cancel).
func (c *SyncController) Stop() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cancelFunc != nil {
		c.cancelFunc()
	}
	c.isActive = false
}

// Cancel fully aborts the current process and triggers OnCancelled.
func (c *SyncController) Cancel() error {
	c.Stop()
	err := c.manager.Cancel()
	c.downloader.Stop()
	
	if c.events.OnCancelled != nil {
		c.events.OnCancelled()
	}
	return err
}

// Status proxies and aggregates reports from Manager and Downloader.
func (c *SyncController) Status() SyncStatusReport {
	report := c.manager.Status()
	
	completed, pending, failed := c.downloader.Progress()
	report.DownloadedChunks = completed
	report.PendingChunks = pending
	report.FailedChunks = failed
	report.Workers = c.downloader.ActiveWorkers()
	// PeerPool is inside Manager
	report.Peers = c.manager.peerPool.PeerCount()
	
	// Assuming blocks downloaded = chunks * chunk size approx
	// In the future this will be precise. For now we use the manager's blocksSynced.
	
	return report
}

func (c *SyncController) changeState(newState SyncState) {
	oldState := c.manager.Status().CurrentState
	c.manager.SetState(newState)
	if c.events.OnStateChanged != nil {
		c.events.OnStateChanged(oldState, newState)
	}
}

// runStateMachine is the core orchestrator.
func (c *SyncController) runStateMachine(ctx context.Context, targetHeight uint64) {
	defer func() {
		c.mu.Lock()
		c.isActive = false
		c.mu.Unlock()
	}()

	// 1. DiscoveringPeers (Mocked logic, assumes peers already in pool)
	c.changeState(StateDiscoveringPeers)
	if c.manager.peerPool.PeerCount() == 0 {
		c.fail(errors.New("no peers available"))
		return
	}

	// 2. RequestingHeaders (Simulated)
	select {
	case <-ctx.Done():
		return
	case <-time.After(50 * time.Millisecond):
	}
	c.changeState(StateRequestingHeaders)

	// Simulate receiving headers and queuing chunks
	// E.g., targetHeight = 100, chunk size = 10 => 10 chunks
	// Actually we should queue based on target height
	chunkSize := uint64(100)
	c.downloader.queue.Reset()
	c.downloader.queue.AddRange(c.manager.localHeight+1, targetHeight, chunkSize)

	// 3. DownloadingBlocks
	c.changeState(StateDownloadingBlocks)
	if c.events.OnDownloadStarted != nil {
		c.events.OnDownloadStarted()
	}
	
	c.downloader.Start()

	// Wait for downloader to finish
	for {
		select {
		case <-ctx.Done():
			c.downloader.Stop()
			return
		case <-time.After(100 * time.Millisecond):
			// Check if downloader is still active and chunks are done
			comp, pend, fail := c.downloader.Progress()
			
			// Emulate height progress based on chunks
			c.manager.UpdateLocalHeight(c.manager.localHeight + uint64(comp*int(chunkSize)))
			
			if fail > 0 {
				c.downloader.Stop()
				c.fail(errors.New("chunk download failed permanently"))
				return
			}
			if pend == 0 && comp > 0 {
				c.downloader.Stop()
				goto verify
			}
			if pend == 0 && comp == 0 {
				c.downloader.Stop()
				goto verify
			}
		}
	}

verify:
	if c.events.OnDownloadCompleted != nil {
		c.events.OnDownloadCompleted()
	}

	// 4. VerifyingBlocks (Simulated check)
	c.changeState(StateVerifyingBlocks)
	select {
	case <-ctx.Done():
		return
	case <-time.After(50 * time.Millisecond):
	}
	
	// Simulate checking continuity (omitted hash/crypto logic as per RFC)
	// 5. Completed
	c.manager.UpdateLocalHeight(targetHeight)
	c.changeState(StateCompleted)
}

func (c *SyncController) fail(err error) {
	c.changeState(StateFailed)
	if c.events.OnFailed != nil {
		c.events.OnFailed(err)
	}
}
