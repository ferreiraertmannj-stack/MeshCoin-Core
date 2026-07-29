package p2p

import (
	"testing"
	"time"
)

func TestSecurityManager_Blacklist(t *testing.T) {
	sm := NewSecurityManager()
	ip := "192.168.1.1"

	if sm.IsBlacklisted(ip) {
		t.Errorf("Expected not blacklisted initially")
	}

	sm.BlacklistPeer(ip, 50*time.Millisecond)

	if !sm.IsBlacklisted(ip) {
		t.Errorf("Expected blacklisted after adding")
	}

	time.Sleep(100 * time.Millisecond)

	if sm.IsBlacklisted(ip) {
		t.Errorf("Expected not blacklisted after expiration")
	}
}

func TestSecurityManager_RegisterNodeID(t *testing.T) {
	sm := NewSecurityManager()

	if !sm.RegisterNodeID("node1") {
		t.Errorf("Expected to register node1 successfully")
	}

	if sm.RegisterNodeID("node1") {
		t.Errorf("Expected to fail registering duplicate node1")
	}

	sm.UnregisterNodeID("node1")

	if !sm.RegisterNodeID("node1") {
		t.Errorf("Expected to register node1 successfully after unregister")
	}
}

func TestSecurityManager_RateLimitConnection(t *testing.T) {
	sm := NewSecurityManager()
	ip := "127.0.0.1"

	// Limit is 5 per 10s
	for i := 0; i < 5; i++ {
		if !sm.AllowConnection(ip) {
			t.Errorf("Expected connection %d to be allowed", i)
		}
	}

	if sm.AllowConnection(ip) {
		t.Errorf("Expected 6th connection to be rate limited")
	}
}

func TestSecurityManager_RateLimitMessage(t *testing.T) {
	sm := NewSecurityManager()
	ip := "10.0.0.1"

	// Limit is 20 per 1s
	for i := 0; i < 20; i++ {
		if !sm.AllowMessage(ip) {
			t.Errorf("Expected message %d to be allowed", i)
		}
	}

	if sm.AllowMessage(ip) {
		t.Errorf("Expected 21st message to be rate limited")
	}
}
