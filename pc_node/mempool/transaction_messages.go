package mempool

// MsgTransaction represents a full transaction over the wire
type MsgTransaction struct {
	Hash      string `json:"hash"`
	Sender    string `json:"sender"`
	Fee       uint64 `json:"fee"`
	Timestamp int64  `json:"timestamp"`
	Payload   []byte `json:"payload"`
}

func (m *MsgTransaction) GetHash() string     { return m.Hash }
func (m *MsgTransaction) GetSender() string   { return m.Sender }
func (m *MsgTransaction) GetFee() uint64      { return m.Fee }
func (m *MsgTransaction) GetSize() uint64     { return uint64(len(m.Payload)) }
func (m *MsgTransaction) GetTimestamp() int64 { return m.Timestamp }
func (m *MsgTransaction) GetNonce() uint64    { return 0 } // Assume decoded from payload if needed, keeping simple

// MsgTransactionAnnouncement represents just the hash (Inventory)
type MsgTransactionAnnouncement struct {
	Hash string `json:"hash"`
}
