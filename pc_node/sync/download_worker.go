package sync

import (
	"context"
	"time"
)

// DownloadedChunk is the payload kept strictly in memory for a downloaded chunk.
type DownloadedChunk struct {
	StartHeight  uint64
	EndHeight    uint64
	Blocks       [][]byte
	PeerID       string
	DownloadTime time.Duration
}

// DownloadWorker is an independent goroutine that asks the queue for chunks
// and uses the best peer to download them.
type DownloadWorker struct {
	id       int
	queue    *DownloadQueue
	peerPool PeerPool
	results  chan<- DownloadedChunk
	timeout  time.Duration
}

// NewDownloadWorker creates a worker bound to a specific queue and pool.
func NewDownloadWorker(id int, queue *DownloadQueue, pool PeerPool, results chan<- DownloadedChunk, timeout time.Duration) *DownloadWorker {
	return &DownloadWorker{
		id:       id,
		queue:    queue,
		peerPool: pool,
		results:  results,
		timeout:  timeout,
	}
}

// Start begins the continuous loop of requesting chunks until context is cancelled.
func (w *DownloadWorker) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			chunk, ok := w.queue.NextChunk()
			if !ok {
				// Queue empty. Wait a bit so we don't spin CPU.
				time.Sleep(10 * time.Millisecond)
				continue
			}

			w.processChunk(ctx, *chunk)
		}
	}
}

// processChunk tries to download the chunk from the best available peer.
func (w *DownloadWorker) processChunk(ctx context.Context, chunk DownloadChunk) {
	peer := w.peerPool.BestPeer()
	if peer == nil {
		// No peers available. Mark failed (requeue) and wait.
		w.queue.MarkFailed(chunk)
		time.Sleep(100 * time.Millisecond)
		return
	}

	start := time.Now()
	errCh := make(chan error, 1)

	// Fire network request asynchronously to respect the timeout
	go func() {
		errCh <- peer.RequestBlocks(chunk.StartHeight, chunk.EndHeight)
	}()

	select {
	case <-ctx.Done():
		w.queue.MarkFailed(chunk)
		return
	case <-time.After(w.timeout):
		// Timeout reached
		peer.AddFailure() // Penalize the peer in the scoring algorithm
		w.queue.MarkFailed(chunk)
	case err := <-errCh:
		if err != nil {
			// Network error or rejection
			peer.AddFailure()
			w.queue.MarkFailed(chunk)
		} else {
			// Success! In this sprint, we just mock the payload in memory.
			w.queue.MarkCompleted(chunk)
			
			w.results <- DownloadedChunk{
				StartHeight:  chunk.StartHeight,
				EndHeight:    chunk.EndHeight,
				Blocks:       make([][]byte, 0), // Kept empty for the mock phase
				PeerID:       peer.ID(),
				DownloadTime: time.Since(start),
			}
		}
	}
}
