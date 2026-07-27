# 05 GLOBAL STATE

## Variáveis Globais Acessadas
- `ledger`: (Struct `Ledger`). Lida e Escrita pelos nós de consenso, protegida por `ledger.mu`.
- `PendingTransactions`: Array em RAM de mempool. Protegida por `mempoolMutex`.
- `ledgerFile`: Path base em string. Modificado para `var` na Sprint 2.0. Lida em I/O.
- `activeTCPClients`: Map de Net.Conn (rede P2P). Protegido por `tcpMutex`.
- `clients`: Map de Net.Conn (WebSockets). Acessado em loop sem Mutex em main.go `handleWebSocket` vs `handleMessages`. **CRÍTICO: Condição de corrida severa de Data Race detectada (main.go linha 114 vs 154).**
- `broadcast`: Chan de envio para WS.
