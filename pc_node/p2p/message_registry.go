package p2p

import (
	"sync"
)

// MessageHandler is the signature for all message handlers
type MessageHandler func(peer *Peer, msg P2PMessage) error

// MessageRegistry maps message types to their handlers
type MessageRegistry struct {
	handlers map[string]MessageHandler
	mu       sync.RWMutex
}

func NewMessageRegistry() *MessageRegistry {
	return &MessageRegistry{
		handlers: make(map[string]MessageHandler),
	}
}

func (r *MessageRegistry) RegisterHandler(msgType string, handler MessageHandler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.handlers[msgType] = handler
}

func (r *MessageRegistry) RemoveHandler(msgType string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.handlers, msgType)
}

func (r *MessageRegistry) GetHandler(msgType string) (MessageHandler, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	handler, ok := r.handlers[msgType]
	return handler, ok
}

func (r *MessageRegistry) HandlerCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.handlers)
}
