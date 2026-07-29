package p2p

import (
	"encoding/json"
	"log"
	"time"
)

func (p *Peer) heartbeatLoop() {
	ticker := time.NewTicker(p.config.HeartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-p.ctx.Done():
			return
		case <-ticker.C:
			// Send Ping
			ping := MsgPing{Timestamp: time.Now().UnixNano()}
			if err := p.SendMessage(MsgTypePing, ping); err != nil {
				p.Disconnect("failed to send ping")
				return
			}

			// Check Timeout
			p.mu.RLock()
			lastPong := p.info.LastPong
			p.mu.RUnlock()

			if !lastPong.IsZero() && time.Since(lastPong) > p.config.HeartbeatTimeout {
				if p.events.OnHeartbeatTimeout != nil {
					p.events.OnHeartbeatTimeout(p.info.NodeID)
				}
				p.Disconnect("heartbeat timeout")
				return
			}
		}
	}
}

func (p *Peer) messageLoop() {
	defer p.Disconnect("message loop ended")

	for {
		select {
		case <-p.ctx.Done():
			return
		default:
		}

		p.conn.SetReadDeadline(time.Now().Add(p.config.HeartbeatTimeout))
		var msg P2PMessage
		if err := p.decoder.Decode(&msg); err != nil {
			return
		}

		// Security: rate limit messages
		ip := p.conn.RemoteAddr().String()
		if !p.secManager.AllowMessage(ip) {
			p.secManager.BlacklistPeer(ip, 5*time.Minute)
			if p.events.OnPeerBlacklisted != nil {
				p.events.OnPeerBlacklisted(ip, "flood protection: too many messages")
			}
			p.Disconnect("flood protection")
			return
		}

		switch msg.Type {
		case MsgTypePing:
			var ping MsgPing
			b, _ := json.Marshal(msg.Payload)
			json.Unmarshal(b, &ping)
			p.SendMessage(MsgTypePong, MsgPong{Timestamp: ping.Timestamp})
			p.mu.Lock()
			p.info.LastPing = time.Now()
			p.mu.Unlock()

		case MsgTypePong:
			var pong MsgPong
			b, _ := json.Marshal(msg.Payload)
			json.Unmarshal(b, &pong)

			rtt := time.Duration(time.Now().UnixNano() - pong.Timestamp)
			p.mu.Lock()
			p.info.Latency = rtt
			p.info.LastPong = time.Now()
			p.mu.Unlock()

		case MsgTypeDisconnect:
			var disc MsgDisconnect
			b, _ := json.Marshal(msg.Payload)
			json.Unmarshal(b, &disc)
			log.Printf("Peer %s disconnected: %s", p.info.NodeID, disc.Reason)
			return

		default:
			// Invalid or unimplemented messages
			// Could be passed down via another callback, but for now we ignore or penalize
		}
	}
}
