package mempool

import (
	"encoding/json"
	"fmt"
)

// P2P Message Types (Assuming constants from P2P module, we use byte aliases here to decouple)
const (
	MsgTypeTransaction byte = 0x15 // Example arbitrary byte
	MsgTypeInventory   byte = 0x16
	MsgTypeGetData     byte = 0x17
	MsgTypeData        byte = 0x18
)

type Router interface {
	RegisterHandler(msgType byte, handler func(peerID string, payload []byte) error)
}

type TransactionHandlers struct {
	pipeline *TransactionValidationPipeline
	queue    *TransactionQueue
	stats    *TransactionNetworkStatistics
	events   TransactionNetworkEvents
}

func NewTransactionHandlers(
	pipeline *TransactionValidationPipeline,
	queue *TransactionQueue,
	stats *TransactionNetworkStatistics,
	events TransactionNetworkEvents,
) *TransactionHandlers {
	return &TransactionHandlers{
		pipeline: pipeline,
		queue:    queue,
		stats:    stats,
		events:   events,
	}
}

func (h *TransactionHandlers) Register(router Router) {
	router.RegisterHandler(MsgTypeTransaction, h.HandleMsgTransaction)
	router.RegisterHandler(MsgTypeInventory, h.HandleMsgInventory)
	router.RegisterHandler(MsgTypeGetData, h.HandleMsgGetData)
	router.RegisterHandler(MsgTypeData, h.HandleMsgData)
}

func (h *TransactionHandlers) HandleMsgTransaction(peerID string, payload []byte) error {
	h.stats.IncReceived()

	var msg MsgTransaction
	err := json.Unmarshal(payload, &msg)
	if err != nil {
		h.stats.IncRejected()
		return fmt.Errorf("failed to decode transaction: %w", err)
	}

	if h.events.OnTransactionReceived != nil {
		go h.events.OnTransactionReceived(msg.Hash)
	}

	return h.queue.Enqueue(&msg)
}

func (h *TransactionHandlers) HandleMsgInventory(peerID string, payload []byte) error {
	// Simple decoding for announcement
	var msg MsgTransactionAnnouncement
	err := json.Unmarshal(payload, &msg)
	if err != nil {
		return err
	}

	h.pipeline.HandleAnnouncement(peerID, msg.Hash)
	return nil
}

func (h *TransactionHandlers) HandleMsgGetData(peerID string, payload []byte) error {
	// Typically the Inventory protocol intercepts this, but if routed directly:
	var msg MsgTransactionAnnouncement
	err := json.Unmarshal(payload, &msg)
	if err != nil {
		return err
	}

	// Fetch from pool
	tx, exists := h.pipeline.pool.Get(msg.Hash)
	if exists {
		// Send back MsgData via Router
		h.stats.IncUploads()

		// This requires router to send data.
		// Realistically handled by NetworkBridge or Inventory.
		raw, _ := json.Marshal(tx)
		_ = h.pipeline.propagator.router.SendToPeer(peerID, MsgTypeData, raw)
	}

	return nil
}

func (h *TransactionHandlers) HandleMsgData(peerID string, payload []byte) error {
	// Incoming requested transaction
	return h.HandleMsgTransaction(peerID, payload)
}
