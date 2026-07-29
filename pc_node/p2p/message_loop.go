package p2p

import (
	"encoding/json"
	"log"
	"time"
)

// MessageLoop is executed in a goroutine for each authenticated peer.
// It receives, decodes, and dispatches messages to the router.
func MessageLoop(peer *Peer, router *MessageRouter) {
	defer peer.Disconnect("message loop ended")

	for {
		select {
		case <-peer.ctx.Done():
			return
		case <-router.ctx.Done():
			return
		default:
		}

		peer.conn.SetReadDeadline(time.Now().Add(peer.config.HeartbeatTimeout))
		var msg P2PMessage
		if err := peer.decoder.Decode(&msg); err != nil {
			return // Disconnect on read error or EOF
		}

		// Security: rate limit messages
		ip := peer.conn.RemoteAddr().String()
		if router.secMgr != nil && !router.secMgr.AllowMessage(ip) {
			router.secMgr.BlacklistPeer(ip, 5*time.Minute)
			if peer.events.OnPeerBlacklisted != nil {
				peer.events.OnPeerBlacklisted(ip, "flood protection: too many messages")
			}
			peer.Disconnect("flood protection")
			return
		}

		// Intercept internal protocol messages (Ping, Pong, Disconnect)
		switch msg.Type {
		case MsgTypePing:
			var ping MsgPing
			b, _ := json.Marshal(msg.Payload)
			json.Unmarshal(b, &ping)
			peer.SendMessage(MsgTypePong, MsgPong{Timestamp: ping.Timestamp})
			peer.mu.Lock()
			info := peer.info
			info.LastPing = time.Now()
			peer.info = info
			peer.mu.Unlock()
			continue

		case MsgTypePong:
			var pong MsgPong
			b, _ := json.Marshal(msg.Payload)
			json.Unmarshal(b, &pong)

			rtt := time.Duration(time.Now().UnixNano() - pong.Timestamp)
			peer.mu.Lock()
			info := peer.info
			info.Latency = rtt
			info.LastPong = time.Now()
			peer.info = info
			peer.mu.Unlock()
			continue

		case MsgTypeDisconnect:
			var disc MsgDisconnect
			b, _ := json.Marshal(msg.Payload)
			json.Unmarshal(b, &disc)
			log.Printf("Peer %s disconnected: %s", peer.Info().NodeID, disc.Reason)
			return
		}

		// Dispatch to the router
		_ = router.Dispatch(peer, msg)
	}
}
