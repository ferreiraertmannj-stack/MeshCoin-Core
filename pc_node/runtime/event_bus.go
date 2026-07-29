package runtime

import (
	"fmt"
	"sync"
)

type Event struct {
	Topic string
	Data  interface{}
}

type EventHandler func(Event)

type EventBus struct {
	mu          sync.RWMutex
	subscribers map[string][]EventHandler
	queue       chan Event
	shutdown    chan struct{}
}

func NewEventBus(queueSize int) *EventBus {
	return &EventBus{
		subscribers: make(map[string][]EventHandler),
		queue:       make(chan Event, queueSize),
		shutdown:    make(chan struct{}),
	}
}

func (b *EventBus) Start() {
	go b.worker()
}

func (b *EventBus) Stop() {
	close(b.shutdown)
}

func (b *EventBus) Subscribe(topic string, handler EventHandler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.subscribers[topic] = append(b.subscribers[topic], handler)
}

func (b *EventBus) Publish(topic string, data interface{}) error {
	evt := Event{Topic: topic, Data: data}
	select {
	case b.queue <- evt:
		return nil
	case <-b.shutdown:
		return fmt.Errorf("event bus stopped")
	default:
		return fmt.Errorf("event bus queue full")
	}
}

func (b *EventBus) worker() {
	for {
		select {
		case <-b.shutdown:
			return
		case evt := <-b.queue:
			b.mu.RLock()
			handlers := b.subscribers[evt.Topic]
			b.mu.RUnlock()

			for _, h := range handlers {
				go h(evt) // dispatch async to prevent blocking the bus
			}
		}
	}
}
