package p2p

import (
	"encoding/json"
	"os"
	"sync"
)

type PeerStore struct {
	mu       sync.RWMutex
	filePath string
	peers    map[string]PeerRecord
}

func NewPeerStore(filePath string) *PeerStore {
	store := &PeerStore{
		filePath: filePath,
		peers:    make(map[string]PeerRecord),
	}
	store.load()
	return store
}

func (s *PeerStore) load() {
	data, err := os.ReadFile(s.filePath)
	if err == nil {
		json.Unmarshal(data, &s.peers)
	}
}

func (s *PeerStore) save() error {
	data, err := json.MarshalIndent(s.peers, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.filePath, data, 0644)
}

func (s *PeerStore) SavePeer(record PeerRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.peers[record.NodeID] = record
	return s.save()
}

func (s *PeerStore) GetPeer(nodeID string) (PeerRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	record, ok := s.peers[nodeID]
	if !ok {
		return record, os.ErrNotExist
	}
	return record, nil
}

func (s *PeerStore) DeletePeer(nodeID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.peers, nodeID)
	return s.save()
}

func (s *PeerStore) GetAllPeers() ([]PeerRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []PeerRecord
	for _, peer := range s.peers {
		result = append(result, peer)
	}
	return result, nil
}
