# IMPLEMENTATION REPORT: RFC-0017-P2P-GOSSIP-PROTOCOL (Nebula Network)

## 1. Arquitetura
A camada de propagação distribuída (Gossip Protocol) da Nebula Network foi integralmente desenhada no pacote `pc_node/p2p`. Para evitar saturação de CPU/Rede e manter o isolamento arquitetural exigido, ela opera como um subsistema plugado ao `MessageRouter`. A arquitetura baseia-se em um modelo "Push Gossip" com backpressure, gerenciando a disseminação por canais assíncronos não-bloqueantes, cache descentralizado com validade temporal e limite de saltos (TTL).

## 2. Gossip Manager
O `GossipManager` é o maestro da propagação. Ele consome as entidades básicas `PeerManager`, `MessageRouter`, coordena as filas (`GossipQueue`) e interroga o `GossipCache`.
Sua principal função de interface é o método estrito genérico de injeção (`Publish()`) utilizado pelo nó local quando minera um bloco ou lança uma transação, e a rotina isolada (`Receive()`) invocada pelos handlers do router ao verem o envelope `GossipEnvelope`.

## 3. Fluxo de Propagação
Quando uma mensagem entra por um Socket (via `Receive()`):
1. **Validação**: Verifica a integridade estrutural (ID existe?). Se falhar: Dropped.
2. **Deduplicação O(1)**: Pergunta ao `GossipCache` se o `MessageID` já existe. Se existir, corta o fluxo (CacheHit), aciona as métricas de redundância e encerra a árvore de broadcast.
3. **TTL Enforcement**: Confere se `TTL > 0`. Se o limite de saltos esgotou, a mensagem morre.
4. **Disparo ao Queue**: O envelope é jogado (sem bloqueio) para a fila `GossipQueue`.
5. **Worker Execution**: Trabalhadores paralelos processam o envio reduzindo o `TTL` e aumentando o `HopCount`.
6. **Encaminhamento Local**: Varre `PeerManager`. **Ignora sempre o nó de onde a mensagem veio (Origin Node)** para não causar loops curtos infinitos, atirando nos demais.

## 4. Cache (GossipCache) e O(1)
Implementado via `map[string]CacheEntry` estritamente guardado por `sync.RWMutex`.
Ao invés de limpar o cache por tamanho restrito (que geraria overhead constante e cache misses prematuros), o cache utiliza base de `time.Duration` global. Uma `goroutine` autônoma (`cleanupLoop`) acorda intermitentemente validando a data de validade de cada entrada e podando o mapa, mantendo o Garbage Collection previsível sem Busy Waits.

## 5. TTL e HopCount
Toda mensagem carrega metadata explícita: `TTL` e `HopCount`.
Ao processar um *Forward*, a lógica decrementa estritamente o `TTL` e avalia se o envio subsequente é digno. Essa é a primeira e fundamental camada contra "Tempestades de Broadcast" (Broadcast Storms), garantindo um alcance finito geográfico no Grafo P2P.

## 6. Backpressure, Retry e Queue
A injeção na rede é feita via `GossipQueue` baseada em `channel`. O *Backpressure* ocorre quando os workers (limitados) não conseguem vazão e a fila lota.
- Nesse cenário de asfixia (QueueFull), a injeção recusa bloqueios, dropa o Forward, e sinaliza a anomalia (QueueOverflow).
- Se a escrita num socket pontual ou processamento interno falhar temporariamente, o `GossipQueue` pode utilizar o campo `Retries`. Uma nova `goroutine` assíncrona agenda a recolocação do pacote com atraso (Backoff simples), salvando mensagens de micro-cortes temporais.

## 7. Eventos Desacoplados
Todas as transições vitais acionam lambdas (`GossipEvents`), rodando invariavelmente via disparos asíncronos (`go m.events...`) sem segurar o fluxo P2P crítico. Eventos contemplados: *Publicação, Recebimento, Forward, Discarding, Duplicação, Queue Overflow, TTL Drop, Propagation Finisehd*.

## 8. Estatísticas Thread-Safe
O subsistema `GossipStatistics` agrega telemetria robusta: Hit Ratio do Cache, Perdas por TTL, Contabilização de Drops e Duplicatas, Medição móvel da Latência do Forward (`AveragePropagationTime`).

## 9. Concorrência Segura
A regra de ouro impera: `Zero Busy Waiting`. Todo loop intermitente (seja o Cache Cleanup ou os Workers da Fila) descansa em `select {}` ouvindo o `context.Done()` do objeto criador. Os acessos aos dicionários são fechados e abertos atomicamente via Lock e Unlock do RWMutex nativo do GO. 

## 10. Testes e Validação
A infraestrutura foi batizada via testes simulando:
- *Stress Test*: 1000 mensagens atiradas a esmo, validadas, deduplicadas e roteadas.
- *Duplicação*: Comprovação de Drops silenciados após o Cache acusar duplicatas.
- *Data Race Check*: Aprovado no detector oficial do `go test -race` atestando robustez no uso dos Mutex e canais.

## 11. Limitações Conhecidas
- *Cache Memory Bound*: Ataques prolongados e massivos gerando centenas de milhares de hashes únicos por segundo em malhas estendidas podem inchar temporariamente a RAM até o TTL agir (poda periódica).

## 12. Próximas Expansões
- *Limitação Estrita de Espaço LRU*: Migrar o Cache apenas baseado em Tempo para LRU com teto rígido (ex: capar em 15.000 entradas dinâmicas).
- *Filtragem Avançada (Bloom Filters)*: Utilizar sumários invertidos e filtros Bloom quando a comunicação inter-peers perguntar pela necessidade de certos `MessageID` antes de trafegar as mensagens grandes (como Mega Blocos).
