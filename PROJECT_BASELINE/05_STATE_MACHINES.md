# 05 STATE MACHINES

## State Machine: TCP Peer Lifecycle
- **Estados:** Disconnected, Handshake (Ping/OGM), Established, Syncing, Banned.
- **Eventos:** Connect, Auth Timeout, Packet Received, Malformed Packet.
- **Falhas:** Timeout sem `OGM` desliga conexão; Ban permanente se PoW falso > 5 vezes.
- **Recuperação:** Auto-reconnect via thread background.

## State Machine: Block Validation
- **Estados:** Receiving, Hash Verification, Tx Verification, Ledger Append, Broadcast.
- **Eventos:** `NEW_BLOCK` payload, ECDSA OK/Fail, PreviousHash Match/Mismatch.
- **Falhas:** Rejeição. Se ledger local maior, envia cadeia superior para Peer.
- **Recuperação:** Drop block. Se fork detectado (mesma altura), armazena pendente.
