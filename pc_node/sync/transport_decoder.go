package sync

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
)

// TransportDecoder decodifica frames TCP
type TransportDecoder struct {
	reader io.Reader
}

func NewTransportDecoder(r io.Reader) *TransportDecoder {
	return &TransportDecoder{reader: r}
}

func (dec *TransportDecoder) Decode() (*TransportMessage, error) {
	// Lemos os 4 bytes do tamanho total
	lenBuf := make([]byte, 4)
	if _, err := io.ReadFull(dec.reader, lenBuf); err != nil {
		return nil, err
	}

	totalLen := binary.BigEndian.Uint32(lenBuf)

	if totalLen < 2 {
		return nil, errors.New("invalid message length")
	}

	if totalLen > MaxPayloadSize+2 {
		return nil, errors.New("message exceeds maximum allowed size")
	}

	// Lemos o resto do frame (Type + Comp + Payload)
	frameBuf := make([]byte, totalLen)
	if _, err := io.ReadFull(dec.reader, frameBuf); err != nil {
		return nil, err
	}

	msgType := MsgType(frameBuf[0])
	compression := frameBuf[1]
	payload := frameBuf[2:]

	// Futuro: se compression != 0x00, descomprimir o payload aqui antes de retornar

	return &TransportMessage{
		Type:        msgType,
		Compression: compression,
		Payload:     payload,
	}, nil
}

// UnmarshalPayload helper para extrair JSON do payload
func (tm *TransportMessage) UnmarshalPayload(v interface{}) error {
	if len(tm.Payload) == 0 {
		return nil // Pode ser mensagem de Ping/Pong
	}
	return json.Unmarshal(tm.Payload, v)
}
