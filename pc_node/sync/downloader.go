package sync

import (
	"context"
	"sync"
	"time"
)

// Downloader orchestrates multiple DownloadWorkers and a DownloadQueue.
type Downloader struct {
	mu         sync.RWMutex
	queue      *DownloadQueue
	peerPool   PeerPool
	workers    []*DownloadWorker
	cancelFunc context.CancelFunc
	results    chan DownloadedChunk

	concurrency int
	timeout     time.Duration
	isActive    bool
	isPaused    bool

	downloadedChunks []DownloadedChunk
}

// NewDownloader creates the parallel download orchestrator.
func NewDownloader(queue *DownloadQueue, pool PeerPool, concurrency int, timeout time.Duration) *Downloader {
	return &Downloader{
		queue:            queue,
		peerPool:         pool,
		concurrency:      concurrency,
		timeout:          timeout,
		results:          make(chan DownloadedChunk, 1000), // Buffer for results
		downloadedChunks: make([]DownloadedChunk, 0),
	}
}

// Start launches the workers as goroutines.
func (d *Downloader) Start() {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.isActive {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	d.cancelFunc = cancel
	d.isActive = true
	d.isPaused = false

	d.workers = make([]*DownloadWorker, d.concurrency)
	for i := 0; i < d.concurrency; i++ {
		worker := NewDownloadWorker(i, d.queue, d.peerPool, d.results, d.timeout)
		d.workers[i] = worker
		go worker.Start(ctx)
	}

	// Collect results async
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case result := <-d.results:
				d.mu.Lock()
				d.downloadedChunks = append(d.downloadedChunks, result)
				d.mu.Unlock()
			}
		}
	}()
}

// Stop sends the cancel signal to all workers and marks downloader as inactive.
func (d *Downloader) Stop() {
	d.mu.Lock()
	defer d.mu.Unlock()
	if !d.isActive {
		return
	}
	if d.cancelFunc != nil {
		d.cancelFunc()
	}
	d.isActive = false
	d.isPaused = false
}

// Pause cancels current workers without losing queue state.
func (d *Downloader) Pause() {
	d.mu.Lock()
	defer d.mu.Unlock()
	if !d.isActive {
		return
	}
	if d.cancelFunc != nil {
		d.cancelFunc()
	}
	d.isPaused = true
}

// Resume restarts workers from where they left off.
func (d *Downloader) Resume() {
	d.mu.Lock()
	defer d.mu.Unlock()
	if !d.isActive || !d.isPaused {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	d.cancelFunc = cancel
	d.isPaused = false

	for i := 0; i < d.concurrency; i++ {
		go d.workers[i].Start(ctx)
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case result := <-d.results:
				d.mu.Lock()
				d.downloadedChunks = append(d.downloadedChunks, result)
				d.mu.Unlock()
			}
		}
	}()
}

// Status returns a textual representation of the downloader state.
func (d *Downloader) Status() string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	if !d.isActive {
		return "Stopped"
	}
	if d.isPaused {
		return "Paused"
	}
	return "Running"
}

// ActiveWorkers returns how many workers are configured to run currently.
func (d *Downloader) ActiveWorkers() int {
	d.mu.RLock()
	defer d.mu.RUnlock()
	if d.isActive && !d.isPaused {
		return d.concurrency
	}
	return 0
}

// Progress returns completed, pending, and failed chunks from the queue.
func (d *Downloader) Progress() (completed, pending, failed int) {
	return d.queue.CompletedChunks(), d.queue.PendingChunks(), d.queue.FailedChunks()
}

// PopDownloadedChunk retrieves and removes the oldest downloaded chunk.
func (d *Downloader) PopDownloadedChunk() (*DownloadedChunk, bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if len(d.downloadedChunks) == 0 {
		return nil, false
	}
	chunk := d.downloadedChunks[0]
	d.downloadedChunks = d.downloadedChunks[1:]
	return &chunk, true
}
