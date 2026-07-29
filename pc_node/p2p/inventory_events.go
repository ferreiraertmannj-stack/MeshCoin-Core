package p2p

type InventoryEvents struct {
	OnInventoryReceived       func(peerID string, items int)
	OnInventoryIgnored        func(peerID string, items int, reason string)
	OnObjectRequested         func(peerID string, objHash string)
	OnObjectReceived          func(peerID string, objHash string)
	OnObjectDelivered         func(objHash string)
	OnObjectTimeout           func(peerID string, objHash string)
	OnObjectNotFound          func(peerID string, objHash string)
	OnQueueOverflow           func()
	OnSynchronizationFinished func(peerID string, syncedItems int)
}
