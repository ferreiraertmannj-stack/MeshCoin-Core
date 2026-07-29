# IMPLEMENTATION REPORT: RFC-0008-DOWNLOADER (Nebula Network)

## 1. Visão Geral (Overview)
A Fase 28 implementa com sucesso o **Pipeline de Download Paralelo** projetado para a arquitetura do Fast Sync. Seu papel é atuar como orquestrador multithread: fracionando o total de blocos atrasados em minúsculos `Chunks`, inserindo-os em uma fila segura (`DownloadQueue`) e delegando o consumo massivo para múltiplos `DownloadWorkers` independentes operando contra o `PeerPool`.

Nenhum arquivo do Consenso, PoW ou Storage original foi modificado nesta fase. Os blocos capturados residem temporariamente em memória na estrutura virtual `DownloadedChunk`.

## 2. Arquivos Impactados
**Arquivos Criados:**
- `pc_node/sync/download_queue.go`: Fila FIFO thread-safe com suporte a retentativas automáticas e contabilidade (pending/completed/failed).
- `pc_node/sync/download_worker.go`: Goroutines independentes e assíncronas dedicadas a requisitar blocos aos Peers.
- `pc_node/sync/downloader.go`: Orquestrador global e gerente do Lifecycle dos Workers, garantindo controle sobre `Start`, `Stop`, `Pause` e `Resume`.
- `pc_node/sync/downloader_test.go`: Suite nativa englobando integração, balanceamento, timeouts e stress testing.

## 3. Arquitetura do Downloader
A orquestração é governada pela struct `Downloader`. Ela opera orientada a eventos via `context.Context`, despachando um pool com tamanho configurável (`concurrency = 16`, etc.). Todos os retornos de sucesso desembocam assincronamente em um Channel (`d.results`), evitando o travamento da malha durante leituras de alto I/O.

## 4. Estratégias do Pipeline
### 4.1. Download Queue
- Emprega um design de array dinâmico (`pendingChunks`) blindado por `sync.RWMutex`.
- Separa o progresso entre `completedChunks` e `failedChunks`.

### 4.2. Estratégia dos Workers
- Consumem o método `NextChunk()` ativamente até que a fila esteja limpa, momento no qual recuam levemente (`time.Sleep`) aguardando encerramento ou nova carga, prevenindo sobrecarga de CPU (Busy-waiting).

### 4.3. Balanceamento de Peers
- Cada Chunk invoca dinamicamente o `BestPeer()` da sua `PeerPool` no exato instante do download. Se o Peer estiver desconectado, o Worker reinjeta o Chunk na fila, penaliza o Score e aguarda.

### 4.4. Política de Retry e Timeout
- **Timeout**: Se a chamada `peer.RequestBlocks()` demorar mais que o tempo configurável injetado via `time.After()`, o Worker aborta a _select-case_, pune o Peer via `AddFailure()` (para não utilizá-lo nas próximas chamadas imediatas) e re-enfileira o Chunk.
- **Retry**: Controlado matematicamente pelo parâmetro `maxRetries` na Queue. Quando o limite cede, o Chunk escoa para `failedChunks`, protegendo a aplicação contra loops infinitos.

## 5. Garantias contra Data Race
- Foi suprimido o uso de mapas desprotegidos.
- Toda modificação nas listas de controle do `Downloader` e da `DownloadQueue` (arrays de chunks pendentes/concluídos) acontece exclusivamente sob as diretrizes limpas de um `mu.Lock()`. O stress test gerou 200 goroutines operando cruzadas contra a malha e confirmou 0 panics de colisão de memória.

## 6. Resultados do Pipeline
Todos os testes terminaram com sucesso absoluto, mantendo a compatibilidade limpa.
- `go fmt ./...`: **Pass**
- `go vet ./...`: **Pass**
- `go test ./...`: **Pass** (Suite aprovada validando recuperação, time-outs paralelos e stress de 200 goroutines)
- `go build ./...`: **Pass**

## 7. Limitações Intencionais
- A struct `DownloadedChunk` guarda atualmente `[][]byte` nulos em memória na validação bem-sucedida, pois o TCP original do projeto não foi espetado. 
- O Storage `Badger` não é importado para evitar a persistência de fato, operando totalmente em isolamento (Sandbox).

## 8. Próximas Expansões Previstas
- Fase 29: Injetar a Máquina de Estados Global (`SyncManager`) para engatilhar o `Downloader` assim que a fase `StateRequestingHeaders` confirmar a integridade.
- Fase 30: Executar a persistência do `DownloadedChunk` contra a abstração Storage.Engine.
