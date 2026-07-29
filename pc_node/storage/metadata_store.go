package storage

import (
	"fmt"
	"sync"
)

type MetadataStore struct {
	mu   sync.RWMutex
	data map[string]interface{}
}

func NewMetadataStore() *MetadataStore {
	return &MetadataStore{
		data: make(map[string]interface{}),
	}
}

func (m *MetadataStore) Save(key string, value interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = value
}

func (m *MetadataStore) Load(key string) (interface{}, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	val, ok := m.data[key]
	if !ok {
		return nil, fmt.Errorf("metadata key %s not found", key)
	}
	return val, nil
}
