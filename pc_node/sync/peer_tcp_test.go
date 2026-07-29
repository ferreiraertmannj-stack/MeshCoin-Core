package sync

import (
	"bytes"
	"net"
	"testing"
	"time"
)

func startMockServer(t *testing.T) net.Listener {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	return l
}

func TestTCPPeer_ConnectAndPing(t *testing.T) {
	l := startMockServer(t)
	defer l.Close()

	go func() {
		conn, err := l.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		dec := NewTransportDecoder(conn)
		msg, _ := dec.Decode()
		if msg.Type == MsgTypePing {
			enc := NewTransportEncoder(conn)
			enc.Encode(MsgTypePong, PongMsg{Timestamp: time.Now()})
		}
	}()

	conn, err := net.Dial("tcp", l.Addr().String())
	if err != nil {
		t.Fatalf("failed to dial: %v", err)
	}

	peer := NewTCPPeer("p1", conn, 0)
	defer peer.Disconnect()

	err = peer.Ping()
	if err != nil {
		t.Fatalf("failed to ping: %v", err)
	}

	msg, err := peer.Receive()
	if err != nil {
		t.Fatalf("failed to receive pong: %v", err)
	}

	if msg.Type != MsgTypePong {
		t.Fatalf("expected MsgTypePong, got %d", msg.Type)
	}
}

func TestTCPPeer_RequestHeaders(t *testing.T) {
	l := startMockServer(t)
	defer l.Close()

	go func() {
		conn, err := l.Accept()
		if err == nil {
			defer conn.Close()
			dec := NewTransportDecoder(conn)
			msg, _ := dec.Decode()
			if msg.Type == MsgTypeRequestHeaders {
				enc := NewTransportEncoder(conn)
				enc.Encode(MsgTypeHeaders, HeadersMsg{
					Headers: []HeaderMetadata{{Hash: "hash1"}},
				})
			}
		}
	}()

	conn, _ := net.Dial("tcp", l.Addr().String())
	peer := NewTCPPeer("p2", conn, 0)
	defer peer.Disconnect()

	err := peer.RequestHeaders(100, 10)
	if err != nil {
		t.Fatalf("RequestHeaders failed: %v", err)
	}

	resp, err := peer.Receive()
	if err != nil {
		t.Fatalf("Receive failed: %v", err)
	}

	if resp.Type != MsgTypeHeaders {
		t.Fatalf("expected headers, got %v", resp.Type)
	}

	var headersMsg HeadersMsg
	if err := resp.UnmarshalPayload(&headersMsg); err != nil {
		t.Fatalf("failed to unmarshal headers: %v", err)
	}

	if len(headersMsg.Headers) != 1 || headersMsg.Headers[0].Hash != "hash1" {
		t.Fatalf("invalid headers content")
	}
}

func TestTCPPeer_RequestBlocks(t *testing.T) {
	l := startMockServer(t)
	defer l.Close()

	go func() {
		conn, err := l.Accept()
		if err == nil {
			defer conn.Close()
			dec := NewTransportDecoder(conn)
			msg, _ := dec.Decode()
			if msg.Type == MsgTypeRequestBlocks {
				enc := NewTransportEncoder(conn)
				enc.Encode(MsgTypeBlocks, BlocksMsg{
					Blocks: [][]byte{[]byte("block1"), []byte("block2")},
				})
			}
		}
	}()

	conn, _ := net.Dial("tcp", l.Addr().String())
	peer := NewTCPPeer("p3", conn, 0)
	defer peer.Disconnect()

	err := peer.RequestBlocks(100, 101)
	if err != nil {
		t.Fatalf("RequestBlocks failed: %v", err)
	}

	resp, err := peer.Receive()
	if err != nil {
		t.Fatalf("Receive failed: %v", err)
	}

	if resp.Type != MsgTypeBlocks {
		t.Fatalf("expected blocks, got %v", resp.Type)
	}

	var blocksMsg BlocksMsg
	if err := resp.UnmarshalPayload(&blocksMsg); err != nil {
		t.Fatalf("failed to unmarshal blocks: %v", err)
	}

	if len(blocksMsg.Blocks) != 2 {
		t.Fatalf("expected 2 blocks, got %d", len(blocksMsg.Blocks))
	}
	if string(blocksMsg.Blocks[0]) != "block1" {
		t.Fatalf("invalid block 0 content")
	}
}

func TestTransportDecoder_InvalidLength(t *testing.T) {
	// Send a payload bigger than MaxPayloadSize
	buf := new(bytes.Buffer)

	// Too big
	header := make([]byte, 6)
	// totalLen > MaxPayloadSize+2
	tooBig := uint32(MaxPayloadSize + 10)
	header[0] = byte(tooBig >> 24)
	header[1] = byte(tooBig >> 16)
	header[2] = byte(tooBig >> 8)
	header[3] = byte(tooBig)

	buf.Write(header)

	dec := NewTransportDecoder(buf)
	_, err := dec.Decode()
	if err == nil {
		t.Fatalf("expected error for oversized payload")
	}
}
