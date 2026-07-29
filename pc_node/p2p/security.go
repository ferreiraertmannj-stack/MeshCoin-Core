package p2p

import (
	"sync"
	"time"
)

// SecurityManager handles rate limiting, blacklisting, and node duplication prevention.
type SecurityManager struct {
	mu           sync.RWMutex
	blacklist    map[string]time.Time
	connectedIDs map[string]bool

	// Rate limiting state
	// To keep it simple, we track connection attempts and generic messages per IP.
	connAttempts map[string][]time.Time
	msgCounts    map[string][]time.Time
}

func NewSecurityManager() *SecurityManager {
	return &SecurityManager{
		blacklist:    make(map[string]time.Time),
		connectedIDs: make(map[string]bool),
		connAttempts: make(map[string][]time.Time),
		msgCounts:    make(map[string][]time.Time),
	}
}

// IsBlacklisted returns true if the IP is currently in the blacklist.
func (sm *SecurityManager) IsBlacklisted(ip string) bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	expiry, exists := sm.blacklist[ip]
	if !exists {
		return false
	}
	if time.Now().After(expiry) {
		return false // Will be cleaned up later, but technically not blacklisted anymore
	}
	return true
}

// BlacklistPeer adds an IP to the blacklist for the specified duration.
func (sm *SecurityManager) BlacklistPeer(ip string, duration time.Duration) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.blacklist[ip] = time.Now().Add(duration)
}

// RegisterNodeID ensures a node ID is only connected once. Returns false if already connected.
func (sm *SecurityManager) RegisterNodeID(nodeID string) bool {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	if sm.connectedIDs[nodeID] {
		return false
	}
	sm.connectedIDs[nodeID] = true
	return true
}

func (sm *SecurityManager) UnregisterNodeID(nodeID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	delete(sm.connectedIDs, nodeID)
}

// AllowConnection applies rate limiting for new connections from an IP.
// Limit: 5 connections per 10 seconds.
func (sm *SecurityManager) AllowConnection(ip string) bool {
	return sm.checkRateLimit(ip, sm.connAttempts, 5, 10*time.Second)
}

// AllowMessage applies rate limiting for general messages/pings from an IP.
// Limit: 20 messages per second.
func (sm *SecurityManager) AllowMessage(ip string) bool {
	return sm.checkRateLimit(ip, sm.msgCounts, 20, time.Second)
}

func (sm *SecurityManager) checkRateLimit(ip string, m map[string][]time.Time, max int, window time.Duration) bool {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	now := time.Now()
	times := m[ip]

	// Filter old attempts
	var recent []time.Time
	for _, t := range times {
		if now.Sub(t) <= window {
			recent = append(recent, t)
		}
	}

	if len(recent) >= max {
		m[ip] = recent
		return false
	}

	recent = append(recent, now)
	m[ip] = recent
	return true
}
