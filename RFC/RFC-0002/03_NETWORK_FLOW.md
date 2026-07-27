# 03 NETWORK FLOW

## Mapeamento de Execução (P2P / Sockets)
1. **Componentes:** `listenUDP`, `broadcastPresence`, `listenTCP`, `startSidecarAPI`.
2. **Ordem e Rotinas:**
   - `startNetwork` invoca goroutines separadas para escuta UDP e TCP.
   - UDP (Porta 5555) recebe pings e varre rede local. `broadcastPresence` joga beacon a cada 5s.
   - TCP (Porta 5556) recebe `Accept()`, spawna `handleConnection` por peer.
   - Sidecar API / WS (Porta 8080) atende Flutter app via `handleWebSocket`.
3. **Locks:** 
   - `tcpMutex` protege o map `activeTCPClients`.
4. **Estruturas Compartilhadas:** `activeTCPClients`, `clients` (websocket), `broadcast` chan.
5. **Condições de Corrida:** Socket TCP lendo/escrevendo depende do `tcpMutex`. A latência de rede retém esse lock por tempos imprevisíveis durante o `.Write()` do `broadcastTCP`.
