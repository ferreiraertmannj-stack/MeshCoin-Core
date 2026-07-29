package mempool

import "fmt"

// GossipProtocol Interface
type GossipProtocol interface {
	Publish(topic string, data []byte) error
}

// RouterProtocol Interface
type RouterProtocol interface {
	SendToPeer(peerID string, msgType byte, payload interface{}) error
}

// InventoryProtocol Interface
type InventoryProtocol interface {
	AnnounceObject(objectType string, hash string) error
	RequestObject(peerID string, objectType string, hash string) error
}

const (
	InvTypeTransaction = "TRANSACTION"
	GossipTopicTx      = "TX_PROPAGATION"
)

type TransactionPropagation struct {
	gossip    GossipProtocol
	inventory InventoryProtocol
	router    RouterProtocol
}

func NewTransactionPropagation(gossip GossipProtocol, inv InventoryProtocol, router RouterProtocol) *TransactionPropagation {
	return &TransactionPropagation{
		gossip:    gossip,
		inventory: inv,
		router:    router,
	}
}

// PropagateNewTransaction is called when a NEW valid transaction hits the local Mempool.
func (p *TransactionPropagation) PropagateNewTransaction(hash string, rawData []byte) error {
	// First announce to Inventory to avoid sending heavy payloads directly
	if p.inventory != nil {
		err := p.inventory.AnnounceObject(InvTypeTransaction, hash)
		if err != nil {
			return fmt.Errorf("failed to announce: %w", err)
		}
	}

	// Then Gossip the payload to those who want it
	if p.gossip != nil {
		err := p.gossip.Publish(GossipTopicTx, rawData)
		if err != nil {
			return fmt.Errorf("failed to gossip: %w", err)
		}
	}

	return nil
}

// RequestTransaction fetches a transaction from a specific peer because we received its announcement.
func (p *TransactionPropagation) RequestTransaction(peerID string, hash string) error {
	if p.inventory != nil {
		return p.inventory.RequestObject(peerID, InvTypeTransaction, hash)
	}
	return nil
}
