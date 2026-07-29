package p2p

const (
	MsgTypeInventory    = "INVENTORY"
	MsgTypeGetData      = "GET_DATA"
	MsgTypeData         = "DATA"
	MsgTypeNotFound     = "NOT_FOUND"
	MsgTypeInventoryAck = "INVENTORY_ACK"
)

type InventoryObjectType string

const (
	InventoryBlock       InventoryObjectType = "BLOCK"
	InventoryTransaction InventoryObjectType = "TRANSACTION"
	InventoryHeader      InventoryObjectType = "HEADER"
)

type InventoryItem struct {
	ObjectID   string              `json:"object_id"`
	ObjectType InventoryObjectType `json:"object_type"`
	ObjectHash string              `json:"object_hash"`
	Size       uint64              `json:"size"`
	Timestamp  int64               `json:"timestamp"`
	OriginNode string              `json:"origin_node"`
}

type MsgInventory struct {
	Items []InventoryItem `json:"items"`
}

type MsgGetData struct {
	Items []InventoryItem `json:"items"`
}

type MsgData struct {
	Item InventoryItem `json:"item"`
	Data []byte        `json:"data"`
}

type MsgNotFound struct {
	Items []InventoryItem `json:"items"`
}

type MsgInventoryAck struct {
	ItemsReceived int `json:"items_received"`
}
