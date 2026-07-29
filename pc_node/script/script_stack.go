package script

import (
	"fmt"
	"sync"
)

type ScriptStack struct {
	elements [][]byte
	mu       sync.RWMutex
	maxDepth int
}

func NewScriptStack(maxDepth int) *ScriptStack {
	return &ScriptStack{
		elements: make([][]byte, 0),
		maxDepth: maxDepth,
	}
}

func (s *ScriptStack) Push(data []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.elements) >= s.maxDepth {
		return fmt.Errorf("stack depth exceeded")
	}

	s.elements = append(s.elements, data)
	return nil
}

func (s *ScriptStack) Pop() ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.elements) == 0 {
		return nil, fmt.Errorf("pop from empty stack")
	}

	index := len(s.elements) - 1
	data := s.elements[index]
	s.elements = s.elements[:index]
	return data, nil
}

func (s *ScriptStack) Peek() ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.elements) == 0 {
		return nil, fmt.Errorf("peek from empty stack")
	}
	return s.elements[len(s.elements)-1], nil
}

func (s *ScriptStack) Depth() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.elements)
}

func (s *ScriptStack) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.elements = make([][]byte, 0)
}
