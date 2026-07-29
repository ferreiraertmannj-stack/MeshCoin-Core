package sync

import (
	"testing"
	"time"
)

type mockPeerPool struct{}

func (m *mockPeerPool) AddPeer(p Peer)       {}
func (m *mockPeerPool) RemovePeer(id string) {}
func (m *mockPeerPool) PeerCount() int       { return 0 }
func (m *mockPeerPool) ListPeers() []Peer    { return nil }
func (m *mockPeerPool) BestPeer() Peer       { return nil }
func (m *mockPeerPool) RandomPeer() Peer     { return nil }
func (m *mockPeerPool) FastestPeer() Peer    { return nil }
func (m *mockPeerPool) HighestPeer() Peer    { return nil }

func TestSyncManager_StateTransitions(t *testing.T) {
	pool := &mockPeerPool{}
	sm := NewSyncManager(pool)

	if sm.Status().CurrentState != StateIdle {
		t.Errorf("Expected Idle state initially, got %v", sm.Status().CurrentState)
	}

	err := sm.StartSync(1000)
	if err != nil {
		t.Fatalf("Failed to start sync: %v", err)
	}

	if sm.Status().CurrentState != StateDiscoveringPeers {
		t.Fatalf("Expected StateDiscoveringPeers, got %v", sm.Status().CurrentState)
	}

	sm.UpdateLocalHeight(500)
	status := sm.Status()
	if status.ProgressPercent != 50.0 {
		t.Errorf("Expected 50%% progress, got %f", status.ProgressPercent)
	}

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
	if sm.Status().CurrentState != StateCompleted {
		t.Errorf("Expected Completed state")
	}

	if err := sm.Cancel(); err != nil {
		t.Errorf("Failed to cancel: %v", err)
	}

	if sm.Status().CurrentState != StateIdle {
		t.Errorf("Expected Idle after cancel, got %v", sm.Status().CurrentState)
	}
}
