package sync

import (
	"testing"
	"time"
)

type mockPeerPool struct{}

func (m *mockPeerPool) GetBestPeer() Peer    { return nil }
func (m *mockPeerPool) GetAllPeers() []Peer  { return nil }
func (m *mockPeerPool) AddPeer(p Peer)       {}
func (m *mockPeerPool) RemovePeer(id string) {}

func TestSyncManager_StateTransitions(t *testing.T) {
	pool := &mockPeerPool{}
	sm := NewSyncManager(pool)

	if sm.Status().State != StateIdle {
		t.Errorf("Expected Idle state initially, got %v", sm.Status().State)
	}

	err := sm.StartSync(1000)
	if err != nil {
		t.Fatalf("Failed to start sync: %v", err)
	}

	if sm.Status().State != StateDiscoveringPeers {
		t.Errorf("Expected DiscoveringPeers state, got %v", sm.Status().State)
	}

	sm.UpdateLocalHeight(500)
	status := sm.Status()
	if status.ProgressPct != 50.0 {
		t.Errorf("Expected 50%% progress, got %f", status.ProgressPct)
	}

	// Sleep slightly to simulate time passing for speed calculation
	time.Sleep(10 * time.Millisecond)
	sm.UpdateLocalHeight(600)

	status2 := sm.Status()
	if status2.SpeedBlocksSec <= 0 {
		t.Errorf("Expected speed > 0, got %f", status2.SpeedBlocksSec)
	}

	if err := sm.Pause(); err != nil {
		t.Errorf("Failed to pause: %v", err)
	}

	if err := sm.Resume(); err != nil {
		t.Errorf("Failed to resume: %v", err)
	}

	sm.SetState(StateCompleted)
	if sm.Status().State != StateCompleted {
		t.Errorf("Expected Completed state")
	}

	if err := sm.Cancel(); err != nil {
		t.Errorf("Failed to cancel: %v", err)
	}

	if sm.Status().State != StateIdle {
		t.Errorf("Expected Idle after cancel, got %v", sm.Status().State)
	}
}
