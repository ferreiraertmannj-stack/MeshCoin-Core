package sync

import (
	"time"
)

// Constantes de limites para proteção contra consumo de memória
const (
	MaxBlocksPerMessage = 500
	MaxPayloadSize      = 10 * 1024 * 1024 // 10MB
)

// Tipos de Mensagem do Protocolo Fast Sync
type MsgType byte

const (
	MsgTypePing           MsgType = 0x01
	MsgTypePong           MsgType = 0x02
	MsgTypeGetStatus      MsgType = 0x03
	MsgTypeStatus         MsgType = 0x04
	MsgTypeRequestHeaders MsgType = 0x05
	MsgTypeHeaders        MsgType = 0x06
	MsgTypeRequestBlocks  MsgType = 0x07
	MsgTypeBlocks         MsgType = 0x08
)

// Estruturas de Mensagens

type TransportMessage struct {
	Type        MsgType
	Compression byte   // 0x00=None, 0x01=Gzip, 0x02=Snappy, 0x03=Zstd (Arquitetura preparada)
	Payload     []byte // Payload serializado
}

type SyncStatusMsg struct {
	Height    uint64
	Genesis   string
	Timestamp int64
}

type GetHeadersMsg struct {
	StartHeight uint64
	Limit       int
}

type HeadersMsg struct {
	Headers []HeaderMetadata
}

type GetBlocksMsg struct {
	StartHeight uint64
	EndHeight   uint64
}

type BlocksMsg struct {
	Blocks [][]byte
}

type PingMsg struct {
	Timestamp time.Time
}

type PongMsg struct {
	Timestamp time.Time
}
