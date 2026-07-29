# IMPLEMENTATION REPORT: RFC-0016-P2P-MESSAGE-ROUTER (Nebula Network)

## 1. Arquitetura
A camada de Roteamento P2P (Message Router) da Nebula Network foi implementada dentro de `pc_node/p2p` mantendo rígido isolamento da lógica de Consenso, Mineração, Ledger e Blockchain. A arquitetura centraliza a recepção, validação, despacho e emissão de mensagens da rede em um `MessageRouter` concorrente que coordena três subcomponentes principais: `MessageRegistry`, `MessageDispatcher` e um Event Loop dedicado.

## 2. Event Loop (Message Loop)
Cada peer autenticado após o Handshake ganha uma goroutine gerenciando o `MessageLoop`. O fluxo:
- Recebe buffer via TCP (`net.Conn`).
- Decodifica o envelope genérico (`P2PMessage`).
- Aplica rate-limiting via `SecurityManager`.
- Intercepta mensagens vitais do protocolo base (`Ping`, `Pong`, `Disconnect`) respondendo nativamente.
- Encaminha mensagens de aplicação via `router.Dispatch(peer, msg)`.
O loop é estritamente atrelado a um `context.Context` (podendo ser encerrado local ou globalmente pelo Router) e sem "busy waits".

## 3. Message Router
O `MessageRouter` funciona como maestro. Ele agrega referências aos gerenciadores (`PeerManager`, `SecurityManager`) sem implementá-los, delegando as funcionalidades específicas. Suas responsabilidades centrais são prover os métodos de `RegisterHandler`, `RemoveHandler`, `Dispatch`, `Broadcast` e `SendToPeer`, suportando encerramento asíncrono e coordenado via `Start()` / `Stop()`.

## 4. Message Dispatcher
O `MessageDispatcher` encapsula a complexidade do callback e segurança de execução:
- Encontra o handler correto para `msg.Type` no Registry.
- Registra o tempo inicial (para cálculo de performance na Estatística).
- Aciona o Handler dentro de uma closure protegida.
- Mede o `duration` (latency) da execução interna e sumariza.
- Propaga callbacks assíncronos (`OnMessageReceived`, `OnMessageDispatched`, `OnUnknownMessage`, `OnDispatchError`).

## 5. Message Registry
Implementa o mapeamento estrito O(1) entre a string de `MessageType` e a função de callback do tipo `MessageHandler func(peer *Peer, msg P2PMessage) error`. Toda escrita (`Register/Remove`) e leitura (`GetHandler`) é protegida por um `sync.RWMutex` prevenindo Data Races.

## 6. Broadcast
O processo de Broadcast (`router.Broadcast(msg)`) varre todos os peers conectados obtidos do `PeerManager`. Para não bloquear ou travar a iteração num peer problemático, o disparo é feito via goroutines independentes (`go func(p *Peer)`), invocando `SendToPeer()`. Se o peer falhar na gravação do socket, o nó aciona a desconexão automática do alvo defeituoso para sanear a topologia.

## 7. Direct Send
`router.SendToPeer(peer, msg)` garante a entrega uni-direcional (Targeted routing). Ele realiza validações prévias de Blacklist (caso o IP já esteja marcado pelo `SecurityManager`), aciona o `peer.SendMessage()` e contabiliza nas `RouterStatistics`.

## 8. Estatísticas
`RouterStatistics` expõe em real-time uma fotografia (Snapshot) concorrentemente segura contendo:
- Total Recebido, Enviado e Despachado
- Total Caiu (Dropped) e Errors (DispatchErrors)
- Total de Broadcasts efetuados
- Handlers simultâneos em execução (`RunningHandlers`)
- Media Móvel Simples (SMA) do Tempo de Despacho (Latência do código da aplicação)

## 9. Recuperação de Panic (Panic Recovery)
Para blindar o Node (imune a crashs causados por Handlers corrompidos), o `MessageDispatcher` invoca a closure da função do Handlers usando blocos de `defer func() { if r := recover(); }`. Em caso de pânico na execução do protocolo de aplicação, o pânico é convertido para `error`, o Router contabiliza o `DispatchError`, aciona o callback assíncrono correspondente e o processo do Nó sobrevive.

## 10. Concorrência
- Não há loops infinitos. Tudo que cicla escuta `context.Done()`.
- O Acesso aos Handlers (Registry), às métricas (Stats) e às rotas (Router) usam extensivamente `sync.RWMutex`.
- O Stress Test atesta operações massivas de escrita, leitura e despacho concorrente (500 goroutines) sem engarrafamentos ou travamentos (Deadlocks).

## 11. Testes e Validação
A suite de testes (`router_test.go`) validou:
1. Registros e Remoções O(1).
2. Queda em Unknown Message quando disparado tipo ausente.
3. Panic Recovery sobrevivendo ao disparo intencional de pânicos.
4. Stress Test: 500 workers disparando paralelamente, com Mutex Locks validando concorrência (Passou no detector de Race do Golang).
5. Broadcast disparado pra simulação de nós e Timeouts.

## 12. Limitações
- O Broadcast atual não usa Gossip Algorithm puro, é um O(N) linear entre conectados. Em cenários de redes com milhões de peers diretos mantidos, o broadcast em malha requererá filas de lote.

## 13. Próximas Expansões
- Adicionar fila de prioridades e backpressure (QoS - Quality of Service) para tratar mensagens vitais (ex: Consenso vs Descoberta) com precedência na rede.
- Adição de criptografia assimétrica ponta-a-ponta (E2EE) caso o payload trafegado requeira privacidade.
