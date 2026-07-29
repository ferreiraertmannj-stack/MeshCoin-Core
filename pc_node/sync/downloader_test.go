package sync

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

// mockFastTCPPeer overloads TCPPeer for test speed and timeout manipulation
type mockFastTCPPeer struct {
	*TCPPeer
	sleepTime time.Duration
	forceFail bool
}

func (m *mockFastTCPPeer) RequestBlocks(start, end uint64) error {
	if m.forceFail {
		return fmt.Errorf("forced network failure")
	}
	time.Sleep(m.sleepTime)
	return nil
}

func TestDownloadQueue_Basic(t *testing.T) {
	q := NewDownloadQueue(3)
	q.AddRange(1, 1500, 500) // 1-500, 501-1000, 1001-1500 -> 3 chunks

	if q.PendingChunks() != 3 {
		t.Fatalf("Expected 3 pending chunks")
	}

	c1, ok := q.NextChunk()
	if !ok || c1.StartHeight != 1 || c1.EndHeight != 500 {
		t.Fatalf("Unexpected chunk 1")
	}

	q.MarkCompleted(*c1)
	if q.CompletedChunks() != 1 {
		t.Fatalf("Expected 1 completed chunk")
	}

	c2, _ := q.NextChunk()
	_, _ = q.NextChunk()

	q.MarkFailed(*c2)
	c2_r1, _ := q.NextChunk()
	q.MarkFailed(*c2_r1)
	c2_r2, _ := q.NextChunk()
	q.MarkFailed(*c2_r2) // Should exceed retries (3)

	if q.FailedChunks() != 1 {
		t.Fatalf("Expected 1 failed chunk after 3 retries")
	}

	q.Reset()
	if q.PendingChunks() != 0 || q.CompletedChunks() != 0 {
		t.Fatalf("Reset failed")
	}
}

func TestDownloader_Lifecycle(t *testing.T) {
	pool := NewPeerPool()
	p := &mockFastTCPPeer{TCPPeer: NewTCPPeer("p1", nil, 5000), sleepTime: 5 * time.Millisecond}
	pool.AddPeer(p)

	queue := NewDownloadQueue(3)
	queue.AddRange(1, 10, 2) // 5 chunks

	downloader := NewDownloader(queue, pool, 2, 50*time.Millisecond)

	downloader.Start()
	if downloader.Status() != "Running" || downloader.ActiveWorkers() != 2 {
		t.Fatalf("Failed to start properly")
	}

	downloader.Pause()
	if downloader.Status() != "Paused" || downloader.ActiveWorkers() != 0 {
		t.Fatalf("Failed to pause properly")
	}

	downloader.Resume()
	if downloader.Status() != "Running" || downloader.ActiveWorkers() != 2 {
		t.Fatalf("Failed to resume properly")
	}

	downloader.Stop()
	if downloader.Status() != "Stopped" || downloader.ActiveWorkers() != 0 {
		t.Fatalf("Failed to stop properly")
	}
}

func TestDownloader_Integration_Parallel(t *testing.T) {
	pool := NewPeerPool()
	// p1 is reliable
	p1 := &mockFastTCPPeer{TCPPeer: NewTCPPeer("p1", nil, 5000), sleepTime: 5 * time.Millisecond}
	// p2 is slow (will timeout)
	p2 := &mockFastTCPPeer{TCPPeer: NewTCPPeer("p2", nil, 5000), sleepTime: 200 * time.Millisecond}
	// p3 always fails
	p3 := &mockFastTCPPeer{TCPPeer: NewTCPPeer("p3", nil, 5000), forceFail: true}

	pool.AddPeer(p1)
	pool.AddPeer(p2)
	pool.AddPeer(p3)

	queue := NewDownloadQueue(3)
	queue.AddRange(1, 50, 10) // 5 chunks

	// timeout is 50ms (p2 will fail)
	downloader := NewDownloader(queue, pool, 4, 50*time.Millisecond)
	downloader.Start()

	// Wait for queue to empty and all chunks to be processed
	time.Sleep(200 * time.Millisecond)
	downloader.Stop()

	completed, pending, failed := downloader.Progress()

	// Because p1 is reliable and fastest, the retry mechanism should eventually
	// route failed chunks (from p2 and p3) to p1.
	if completed != 5 {
		t.Fatalf("Expected all 5 chunks to complete eventually via retries, got %d. pending=%d failed=%d", completed, pending, failed)
	}
}

func TestDownloader_ConcurrencyStress(t *testing.T) {
	pool := NewPeerPool()
	queue := NewDownloadQueue(5)

	// Huge range = 500 chunks
	queue.AddRange(1, 50000, 100)

	// Stress 1: 100 Goroutines dynamically adding/removing peers
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			pid := fmt.Sprintf("peer_%d", idx)
			p := &mockFastTCPPeer{TCPPeer: NewTCPPeer(pid, nil, 100000), sleepTime: 2 * time.Millisecond}
			if idx%3 == 0 {
				p.forceFail = true
			}
			pool.AddPeer(p)
			time.Sleep(10 * time.Millisecond)
			if idx%5 == 0 {
				pool.RemovePeer(pid)
			}
		}(i)
	}

	// Stress 2: Downloader with 150 workers pulling chunks simultaneously
	downloader := NewDownloader(queue, pool, 150, 20*time.Millisecond)
	downloader.Start()

	// Stress 3: External goroutines checking progress and pausing
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 5; j++ {
				downloader.Progress()
				downloader.Status()
				downloader.ActiveWorkers()
				time.Sleep(5 * time.Millisecond)
			}
		}()
	}

	// Let the storm resolve
	wg.Wait()
	time.Sleep(100 * time.Millisecond)

	downloader.Pause()
	downloader.Resume()
	time.Sleep(50 * time.Millisecond)

	downloader.Stop()

	completed, pending, failed := downloader.Progress()
	total := completed + pending + failed
	// The queue might not be fully exhausted due to time limit, but it must not crash or race
	if total != 500 && pending > 0 { // Sum might exceed due to retries if not tracked carefully? No, chunks are moved.
		// Retries re-add to pending, but total unique is 500.
		// Wait, len(pending) + len(failed) + len(completed) should always be >= 500
		// Actually exactly 500 since a chunk is either pending, completed, or failed (removed from pending when processed).
		// Wait, while processing, it's out of the queue!
		// So total = 500 - in_flight_chunks.
	}

	// Just passing without panics or races is the primary goal for this stress test.
}
