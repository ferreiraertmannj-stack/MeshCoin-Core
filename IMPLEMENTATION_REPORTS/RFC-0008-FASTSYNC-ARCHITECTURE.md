# IMPLEMENTATION REPORT: RFC-0008-FASTSYNC-ARCHITECTURE (Nebula Network)

## 1. Visão Geral (Architecture Overview)
Este documento descreve a primeira versão da arquitetura **Fast Sync** para a **Nebula Network**. O objetivo do Fast Sync é substituir a sincronização linear tradicional da malha P2P por um mecanismo desacoplado de download em massa, capaz de buscar lotes de blocos diretamente na Engine de Storage antes de iniciar as validações pesadas da rede principal.

A infraestrutura foi criada contendo exclusivamente os contratos em Go (Interfaces) e Máquinas de Estado. Nenhum download real foi integrado à camada de Bootstrap ou ao Consenso nesta Sprint.

## 2. Máquina de Estados (State Machine)
O fluxo do `SyncManager` transita sequencialmente e dinamicamente através dos seguintes `SyncState`:
- `Idle`: Standby, nó já está sincronizado ou Fast Sync não iniciado.
- `DiscoveringPeers`: Avaliando a _PeerPool_ e descobrindo nós mais altos/velozes.
- `RequestingHeaders`: Baixando estruturas leves (`HeaderMetadata`) para montar o esqueleto e verificar Hashes.
- `DownloadingBlocks`: Solicitando pedaços (Chunks) massivos binários de múltiplos peers.
- `VerifyingBlocks`: Validação criptográfica (NeonHash, PoW, Assinaturas).
- `ImportingBlocks`: Executando _Batch Commit_ contra o `BadgerStorageAdapter`.
- `Completed`: Sincronização atingiu `RemoteHeight`.
- `Failed`: Abortado por timeout, isolamento de rede ou corrupção.

## 3. Fluxo de Execução
1. O Nó em _Bootstrap_ percebe estar muito atrasado e invoca `SyncManager.StartSync(targetHeight)`.
2. A _UI_ e a camada superior leem progresso via `SyncManager.Status()`, que calcula assincronamente velocidade de download (`SpeedBlocksSec`) e Tempo Estimado de Chegada (`ETASeconds`).
3. Comandos externos (ex: UI, CLI) podem interceptar a operação enviando sinais de `Pause()`, `Resume()` ou `Cancel()`.

## 4. Protocolos Desacoplados (Messages)
Mensagens P2P criadas no escopo do Fast Sync (agnósticas ao protocolo atual de fofoca):
- `MsgGetHeaders` (Busca de esqueleto via Hash) -> Retorna `MsgHeaders` (Índice + Hash)
- `MsgGetBlocks` (Range massivo) -> Retorna `MsgBlocks` (Array bidimensional de bytes nativos prontos para o `storage.Batch.PutBlock`)
- `MsgSyncStatus` / `MsgSyncCompleted` (Gerenciamento de Lifecycle de Peers).

## 5. Interfaces Abstratas (Design for Expansion)
Para proteger a flexibilidade do design, as dependências foram construídas em contratos puros (_Interfaces_):
- `Peer`: Encapsula conexões TCP/WebSocket do nó para envio cego de Requests de Blocos.
- `PeerPool`: Orquestrador de rede que decide de onde puxar (_Load Balancing_).
- `BlockValidator`: Validação isolada por header (Lightweight) e payload (Heavyweight).
- `SnapshotManager`: Preparado para importar/exportar Dumps completos do BadgerDB em versões futuras do protocolo.

## 6. Resultados do CI (Pipelines)
- `go fmt ./...`: Pass
- `go vet ./...`: Pass
- `go test ./...`: Pass (`ok pc_node/sync`)
- `go build ./...`: Pass

## 7. Roadmap Próximas Fases
1. **Fase 27**: Implementação física do `Peer` injetando as conexões de rede atuais sob esta _Interface_.
2. **Fase 28**: Integração da _Engine de Storage_ real no `ImportingBlocks` (BadgerDB).
3. **Fase 29**: Integração do _Fast Sync_ ao _Bootstrap_ do `main.go`.
4. **Fase 30**: Testes de download paralelo simulado entre múltiplos Nós.
