package sync

import (
	"sync"
	"testing"
	"time"
)

func TestSyncController_StateTransitions(t *testing.T) {
	pool := NewPeerPool()
	p := &mockFastTCPPeer{TCPPeer: NewTCPPeer("p1", nil, 5000), sleepTime: 1 * time.Millisecond}
	pool.AddPeer(p)

	manager := NewSyncManager(pool)
	queue := NewDownloadQueue(3)
	downloader := NewDownloader(queue, pool, 2, 50*time.Millisecond)
	
	stateChanges := []SyncState{}
	var stateMu sync.Mutex

	events := SyncControllerEventHandlers{
		OnStateChanged: func(oldState, newState SyncState) {
			stateMu.Lock()
			stateChanges = append(stateChanges, newState)
			stateMu.Unlock()
		},
		OnDownloadStarted:   func() {},
		OnDownloadCompleted: func() {},
	}

	controller := NewSyncController(manager, downloader, events)

	err := controller.Start(100)
	if err != nil {
		t.Fatalf("Failed to start controller: %v", err)
	}

	// Wait for states to pass
	time.Sleep(300 * time.Millisecond)

	status := controller.Status()
	if status.CurrentState != StateCompleted {
		t.Fatalf("Expected StateCompleted, got %v", status.CurrentState)
	}

	stateMu.Lock()
	if len(stateChanges) < 4 {
		t.Fatalf("Not enough state transitions logged: %v", stateChanges)
	}
	stateMu.Unlock()
}

func TestSyncController_CancelAndPause(t *testing.T) {
	pool := NewPeerPool()
	p := &mockFastTCPPeer{TCPPeer: NewTCPPeer("p1", nil, 5000), sleepTime: 10 * time.Millisecond}
	pool.AddPeer(p)

	manager := NewSyncManager(pool)
	queue := NewDownloadQueue(3)
	downloader := NewDownloader(queue, pool, 2, 50*time.Millisecond)
	
	cancelled := false
	events := SyncControllerEventHandlers{
		OnCancelled: func() { cancelled = true },
	}

	controller := NewSyncController(manager, downloader, events)
	
	controller.Start(5000)
	time.Sleep(10 * time.Millisecond)
	
	controller.Pause()
	status := controller.Status()
	if status.CurrentState == StateCompleted {
		t.Fatalf("Should not be completed")
	}

	controller.Resume()
	
	controller.Cancel()
	time.Sleep(10 * time.Millisecond)
	
	if !cancelled {
		t.Fatalf("OnCancelled event not fired")
	}
	if controller.Status().CurrentState != StateIdle {
		t.Fatalf("Expected StateIdle after cancel, got %v", controller.Status().CurrentState)
	}
}

func TestSyncController_ConcurrencyStress(t *testing.T) {
	pool := NewPeerPool()
	p := &mockFastTCPPeer{TCPPeer: NewTCPPeer("p1", nil, 5000), sleepTime: 1 * time.Millisecond}
	pool.AddPeer(p)

	manager := NewSyncManager(pool)
	queue := NewDownloadQueue(3)
	downloader := NewDownloader(queue, pool, 10, 50*time.Millisecond)
	
	controller := NewSyncController(manager, downloader, SyncControllerEventHandlers{})
	controller.Start(10000)
	
	var wg sync.WaitGroup
	for i := 0; i < 300; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			
			// Mix actions
			action := idx % 5
			switch action {
			case 0:
				controller.Status()
			case 1:
				controller.Pause()
				time.Sleep(1 * time.Millisecond)
				controller.Resume()
			case 2:
				_ = controller.Status().ETASeconds
			case 3:
				_ = controller.Status().SpeedBlocksSec
			case 4:
				if idx == 299 { // Only one cancel at the end
					time.Sleep(50 * time.Millisecond)
					controller.Cancel()
				}
			}
		}(i)
	}

	wg.Wait()
	// Must not race or panic.
}
