package p2p

import (
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
