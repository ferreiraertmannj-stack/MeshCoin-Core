package p2p

const (
	MsgTypeGetHeaders = "GET_HEADERS"
	MsgTypeHeaders    = "HEADERS"
	MsgTypeGetBlocks  = "GET_BLOCKS"
)

type MsgGetHeaders struct {
	LocatorHashes []string `json:"locator_hashes"`
	HashStop      string   `json:"hash_stop"`
}

type BlockHeader struct {
	Hash       string `json:"hash"`
	ParentHash string `json:"parent_hash"`
	Height     uint64 `json:"height"`
	// Outros metadados omitidos por simplicidade (merkle root, nonce, etc)
}

type MsgHeaders struct {
	Headers []BlockHeader `json:"headers"`
}

type MsgGetBlocks struct {
	LocatorHashes []string `json:"locator_hashes"`
	HashStop      string   `json:"hash_stop"`
}
