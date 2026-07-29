package sync

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
)

// Encoder abstrai a serialização de mensagens e envio no socket TCP
type TransportEncoder struct {
	writer io.Writer
}

func NewTransportEncoder(w io.Writer) *TransportEncoder {
	return &TransportEncoder{writer: w}
}

func (enc *TransportEncoder) Encode(msgType MsgType, payload interface{}) error {
	var payloadBytes []byte
	var err error

	if payload != nil {
		payloadBytes, err = json.Marshal(payload)
		if err != nil {
			return err
		}
	}

	if len(payloadBytes) > MaxPayloadSize {
		return errors.New("payload exceeds maximum allowed size")
	}

	// Frame da mensagem:
	// 4 bytes: Tamanho total (Type + Comp + Payload)
	// 1 byte: Tipo da mensagem
	// 1 byte: Compressão
	// N bytes: Payload

	totalLen := 2 + len(payloadBytes)

	header := make([]byte, 6)
	binary.BigEndian.PutUint32(header[0:4], uint32(totalLen))
	header[4] = byte(msgType)
	header[5] = 0x00 // Nenhuma compressão ativada (mas arquitetura preparada)

	// Escrever header
	if _, err := enc.writer.Write(header); err != nil {
		return err
	}

	// Escrever payload
	if len(payloadBytes) > 0 {
		if _, err := enc.writer.Write(payloadBytes); err != nil {
			return err
		}
	}

	return nil
}
