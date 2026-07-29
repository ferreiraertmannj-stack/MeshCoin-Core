package sync

import (
	"context"
	"errors"
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

	// Pipeline events
	OnChunkDownloaded func(chunk DownloadedChunk)
	OnChunkValidated  func(chunk DownloadedChunk)
	OnChunkImported   func(chunk DownloadedChunk)
	OnValidationError func(err error, chunk DownloadedChunk)
	OnProgress        func(report SyncStatusReport)
	OnPipelineError   func(err error)
	OnCompleted       func(report SyncStatusReport)
}

// SyncController binds the SyncManager, Downloader, Validator, and Importer.
type SyncController struct {
	mu         sync.RWMutex
	manager    *SyncManager
	downloader *Downloader
	validator  *StandardBlockValidator
	importer   *BlockImporter
	events     SyncControllerEventHandlers

	cancelFunc context.CancelFunc
	isActive   bool
}

// NewSyncController initializes the main coordinator for Fast Sync pipeline.
func NewSyncController(
	manager *SyncManager,
	downloader *Downloader,
	validator *StandardBlockValidator,
	importer *BlockImporter,
	events SyncControllerEventHandlers,
) *SyncController {
	return &SyncController{
		manager:    manager,
		downloader: downloader,
		validator:  validator,
		importer:   importer,
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

// Pause pauses the pipeline.
func (c *SyncController) Pause() error {
	err := c.manager.Pause()
	if err == nil {
		c.downloader.Pause()
	}
	return err
}

// Resume restarts the pipeline.
func (c *SyncController) Resume() error {
	err := c.manager.Resume()
	if err == nil {
		c.downloader.Resume()
	}
	return err
}

// Stop cleanly aborts the sync without emitting a failure.
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

// Status aggregates reports from Manager, Downloader, Validator, and Importer.
func (c *SyncController) Status() SyncStatusReport {
	report := c.manager.Status()

	comp, pend, fail := c.downloader.Progress()
	report.DownloadedChunks = comp
	report.PendingChunks = pend
	report.FailedChunks = fail
	report.Workers = c.downloader.ActiveWorkers()
	report.Peers = c.manager.peerPool.PeerCount()

	// Validator Stats
	vStats := c.validator.Statistics()
	report.ValidatedBlocks = vStats.BlocksValidated
	report.RejectedBlocks = vStats.BlocksRejected
	report.BytesDownloaded = vStats.BytesValidated // As proxy

	// Importer Stats
	iStats := c.importer.Statistics()
	report.ImportedBlocks = iStats.ImportedBlocks
	report.BytesImported = iStats.ImportedBytes
	report.ChunksProcessed = int(iStats.ImportedChunks)

	// Combine to overall progress
	report.DownloadedBlocks = report.ImportedBlocks // Using imported as finalized

	if report.RemoteHeight > 0 {
		report.ProgressPercent = (float64(report.ImportedBlocks) / float64(report.RemoteHeight)) * 100
	}

	report.SpeedBlocksSec = iStats.ImportSpeed
	report.ElapsedTime = iStats.ElapsedTime

	if report.SpeedBlocksSec > 0 && report.RemoteHeight > report.CurrentHeight {
		report.ETASeconds = float64(report.RemoteHeight-report.CurrentHeight) / report.SpeedBlocksSec
	}

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

	c.changeState(StateDiscoveringPeers)
	if c.manager.peerPool.PeerCount() == 0 {
		c.fail(errors.New("no peers available"))
		return
	}

	select {
	case <-ctx.Done():
		return
	case <-time.After(10 * time.Millisecond):
	}
	c.changeState(StateRequestingHeaders)

	peer := c.manager.peerPool.BestPeer()
	if peer == nil {
		c.fail(errors.New("no peers available for headers"))
		return
	}

	err := peer.RequestHeaders(c.manager.localHeight+1, int(targetHeight-c.manager.localHeight))
	if err != nil {
		c.fail(err)
		return
	}

	msg, err := peer.Receive()
	if err != nil || msg.Type != MsgTypeHeaders {
		c.fail(errors.New("failed to receive headers"))
		return
	}

	var headersMsg HeadersMsg
	if err := msg.UnmarshalPayload(&headersMsg); err != nil {
		c.fail(err)
		return
	}

	// Montar DownloadQueue com base na quantidade real reportada/recebida
	chunkSize := uint64(100)
	c.downloader.queue.Reset()

	// Apenas como simplificação para a fase, adicionamos range matemático até o targetHeight.
	// O real seria iterar sobre headersMsg.Headers.
	c.downloader.queue.AddRange(c.manager.localHeight+1, targetHeight, chunkSize)

	c.changeState(StateDownloadingBlocks)
	if c.events.OnDownloadStarted != nil {
		c.events.OnDownloadStarted()
	}

	c.downloader.Start()

	// Pipeline loop
	ticker := time.NewTicker(5 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			c.downloader.Stop()
			return
		case <-ticker.C:
			// Process downloaded chunks
			for {
				chunk, ok := c.downloader.PopDownloadedChunk()
				if !ok {
					break // No more chunks right now
				}

				if c.events.OnChunkDownloaded != nil {
					c.events.OnChunkDownloaded(*chunk)
				}

				c.changeState(StateVerifyingBlocks)
				if err := c.validator.ValidateChunk(*chunk); err != nil {
					// Disparar OnValidationError, não abortar o pipeline, descartar chunk
					if c.events.OnValidationError != nil {
						c.events.OnValidationError(err, *chunk)
					}
					continue
				}

				if c.events.OnChunkValidated != nil {
					c.events.OnChunkValidated(*chunk)
				}

				c.changeState(StateImportingBlocks)
				if err := c.importer.ImportChunk(*chunk); err != nil {
					c.fail(err)
					c.downloader.Stop()
					return
				}

				if c.events.OnChunkImported != nil {
					c.events.OnChunkImported(*chunk)
				}

				c.manager.UpdateLocalHeight(chunk.EndHeight)
				if c.events.OnProgress != nil {
					c.events.OnProgress(c.Status())
				}
			}

			// Check completion or downloader failure
			comp, pend, fail := c.downloader.Progress()
			if fail > 0 {
				err := errors.New("pipeline error: chunks permanently failed download")
				if c.events.OnPipelineError != nil {
					c.events.OnPipelineError(err)
				}
				c.downloader.Stop()
				c.fail(err)
				return
			}

			// Finished
			if pend == 0 && comp > 0 {
				// All chunks downloaded and processed?
				// Progress() shows what downloader thinks is completed.
				// Since we popped all downloaded chunks, if queue is done, we are done.
				c.downloader.Stop()
				c.changeState(StateCompleted)
				if c.events.OnDownloadCompleted != nil {
					c.events.OnDownloadCompleted()
				}
				if c.events.OnCompleted != nil {
					c.events.OnCompleted(c.Status())
				}
				return
			}
			if pend == 0 && comp == 0 {
				c.downloader.Stop()
				c.changeState(StateCompleted)
				if c.events.OnCompleted != nil {
					c.events.OnCompleted(c.Status())
				}
				return
			}
		}
	}
}

func (c *SyncController) fail(err error) {
	c.changeState(StateFailed)
	if c.events.OnFailed != nil {
		c.events.OnFailed(err)
	}
}
