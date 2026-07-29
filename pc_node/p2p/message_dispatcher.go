package p2p

import (
	"fmt"
	"time"
)

type MessageDispatcher struct {
	registry *MessageRegistry
	stats    *RouterStatistics
	events   RouterEvents
}

func NewMessageDispatcher(registry *MessageRegistry, stats *RouterStatistics, events RouterEvents) *MessageDispatcher {
	return &MessageDispatcher{
		registry: registry,
		stats:    stats,
		events:   events,
	}
}

// Dispatch routes the message to the correct handler, wrapped in a panic recovery
func (d *MessageDispatcher) Dispatch(peer *Peer, msg P2PMessage) error {
	peerID := ""
	if peer != nil {
		peerID = peer.Info().NodeID
	}

	if d.events.OnMessageReceived != nil {
		go d.events.OnMessageReceived(peerID, msg.Type)
	}

	d.stats.AddReceived()

	handler, ok := d.registry.GetHandler(msg.Type)
	if !ok {
		d.stats.AddUnknown()
		if d.events.OnUnknownMessage != nil {
			go d.events.OnUnknownMessage(peerID, msg.Type)
		}
		return fmt.Errorf("unknown message type: %s", msg.Type)
	}

	d.stats.IncRunning()
	defer d.stats.DecRunning()

	start := time.Now()

	// Panic Recovery block
	var err error
	func() {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("panic in message handler for %s: %v", msg.Type, r)
			}
		}()
		err = handler(peer, msg)
	}()

	duration := time.Since(start)

	if err != nil {
		d.stats.AddDispatchError()
		if d.events.OnDispatchError != nil {
			go d.events.OnDispatchError(peerID, msg.Type, err)
		}
		return err
	}

	d.stats.AddDispatched(duration)
	if d.events.OnMessageDispatched != nil {
		go d.events.OnMessageDispatched(peerID, msg.Type, duration)
	}

	return nil
}
