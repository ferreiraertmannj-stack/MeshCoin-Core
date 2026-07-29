# IMPLEMENTATION REPORT: RFC-0008-PEER-TRANSPORT (Nebula Network)

## 1. Visão Geral (Overview)
A Fase 27 conclui o encapsulamento e a gerência física da estrutura conectiva do **Fast Sync**. A responsabilidade recai inteiramente sobre a orquestração segura de conexões remotas usando Mutex e modelos determinísticos de Score (pontuação) para tomada de decisão (qual `Peer` deve entregar blocos ao nosso Nó validador local).

## 2. Arquivos Impactados
**Arquivos Criados:**
- `pc_node/sync/peer_tcp.go`: Implementação física do pacote.
- `pc_node/sync/peer_pool.go`: _State-Manager_ thread-safe de múltiplos Nodes.
- `pc_node/sync/peer_pool_test.go`: Suite nativa de concorrência e Score.

**Arquivos Alterados:**
- `pc_node/sync/peer_sync.go`: Assinatura original do `PeerPool` e `Peer` evoluiu para incluir os métodos solicitados (_Count, Highest, Best_, propriedades temporais, etc).
- `pc_node/sync/sync_test.go`: Acomodou os mockups da expansão das novas assinaturas do `PeerPool`.

## 3. Interfaces Implementadas
- A interface estrita `Peer` agora é suportada nativamente pela estrutura `TCPPeer` protegendo e encobrindo o ponteiro do `net.Conn`. Nenhuma outra camada do Node da Nebula Network terá acesso ao Socket TCP cru por esse pipeline.
- A interface `PeerPool` é suportada nativamente pela `DefaultPeerPool` estruturada com _RWMutex_.

## 4. Estratégia de Seleção de Peers
Implementada de maneira centralizada via método `calculateScore` injetado e acionado pelo `BestPeer()`. O algoritmo obedece rigidamente à fórmula matemática estipulada pela RFC:
`Score = (height × 4) - latency(ms) - (failures × 100) + connectionTimeBonus(seconds)`

Os testes unitários provaram (via `p1` e `p2` no `peer_pool_test.go`) que essa heurística escolhe deterministicamente e degrada de forma correta quando o `Peer` falha em entregar blocos. Outras funções pontuais operam como seleções atômicas: `RandomPeer()`, `FastestPeer()` e `HighestPeer()`.

## 5. Garantias contra Data Race
A alocação e remoção da estrutura é unicamente conduzida com a trava `sync.RWMutex`.
- `RLock()` é emitido para `PeerCount()`, `ListPeers()`, `BestPeer()`, `HighestPeer()`, permitindo 100% de concorrência massiva de leitura durante a mineração e sincronização simultânea.
- `Lock()` é emitido para `AddPeer()` e `RemovePeer()`.

O pipeline superou o teste de stress que dispara e resolve `Add/Remove/List` via **100 goroutines** agindo imediatamente sob o cache da mesma instância do Pool.

## 6. Resultados do Pipeline
Todos os testes terminaram com sucesso absoluto, mantendo a compatibilidade limpa.
- `go fmt ./...`: **Pass**
- `go vet ./...`: **Pass**
- `go test ./...`: **Pass** (`ok pc_node/sync`)
- `go build ./...`: **Pass**

## 7. Limitações Intencionais
Respeitando as ordens estritas:
- Nenhuma validação foi injetada.
- Nenhuma chamada P2P atual do pacote `protocol` ou pacote base de redes foi violada.
- Não existem processos do Storage (_Badger_) lendo os Chunks.
- `SendMsg` opera perfeitamente abstraído para aceitar serializações Go ou JSON futuramente na camada global do Node, evitando violar as diretrizes de escopo contido.

## 8. Próximas Expansões Previstas
- O próximo gatilho no _Fast Sync_ permitirá que a arquitetura chame diretamente o BadgerStorageAdapter para despejar blocos em Batch assim que o `TCPPeer` decodificar o Payload na rede.
- Ancoragem definitiva do Node (no arquivo `main.go`) passando sua fábrica nativa de conexões para a injeção via `DefaultPeerPool.AddPeer()`.
