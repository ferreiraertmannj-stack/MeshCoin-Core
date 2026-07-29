package sync

import (
	"sync"
)

// DownloadChunk represents a range of blocks to be downloaded.
type DownloadChunk struct {
	StartHeight uint64
	EndHeight   uint64
	Retries     int
}

// DownloadQueue manages the distribution of chunks to workers.
// It is fully thread-safe and supports retries and failure tracking.
type DownloadQueue struct {
	mu              sync.RWMutex
	pendingChunks   []DownloadChunk
	completedChunks []DownloadChunk
	failedChunks    []DownloadChunk
	maxRetries      int
}

// NewDownloadQueue creates a new empty queue with a specific retry limit.
func NewDownloadQueue(maxRetries int) *DownloadQueue {
	return &DownloadQueue{
		pendingChunks:   make([]DownloadChunk, 0),
		completedChunks: make([]DownloadChunk, 0),
		failedChunks:    make([]DownloadChunk, 0),
		maxRetries:      maxRetries,
	}
}

// AddRange splits a large range into smaller chunks and queues them.
func (q *DownloadQueue) AddRange(start, end, chunkSize uint64) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if start > end {
		return
	}

	for i := start; i <= end; i += chunkSize {
		e := i + chunkSize - 1
		if e > end {
			e = end
		}
		q.pendingChunks = append(q.pendingChunks, DownloadChunk{
			StartHeight: i,
			EndHeight:   e,
			Retries:     0,
		})
	}
}

// NextChunk provides the next available chunk for a worker.
func (q *DownloadQueue) NextChunk() (*DownloadChunk, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.pendingChunks) == 0 {
		return nil, false
	}

	chunk := q.pendingChunks[0]
	q.pendingChunks = q.pendingChunks[1:]
	return &chunk, true
}

// MarkCompleted flags a chunk as successfully downloaded.
func (q *DownloadQueue) MarkCompleted(chunk DownloadChunk) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.completedChunks = append(q.completedChunks, chunk)
}

// MarkFailed increments the retry counter. If it exceeds maxRetries,
// the chunk is permanently failed. Otherwise, it is requeued.
func (q *DownloadQueue) MarkFailed(chunk DownloadChunk) {
	q.mu.Lock()
	defer q.mu.Unlock()

	chunk.Retries++
	if chunk.Retries >= q.maxRetries {
		q.failedChunks = append(q.failedChunks, chunk)
	} else {
		// Requeue for retry
		q.pendingChunks = append(q.pendingChunks, chunk)
	}
}

// PendingChunks returns the number of chunks waiting to be downloaded.
func (q *DownloadQueue) PendingChunks() int {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return len(q.pendingChunks)
}

// CompletedChunks returns the number of successfully downloaded chunks.
func (q *DownloadQueue) CompletedChunks() int {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return len(q.completedChunks)
}

// FailedChunks returns the number of permanently failed chunks.
func (q *DownloadQueue) FailedChunks() int {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return len(q.failedChunks)
}

// Reset clears the queue entirely.
func (q *DownloadQueue) Reset() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.pendingChunks = make([]DownloadChunk, 0)
	q.completedChunks = make([]DownloadChunk, 0)
	q.failedChunks = make([]DownloadChunk, 0)
}
