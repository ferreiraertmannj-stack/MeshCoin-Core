# IMPLEMENTATION REPORT: RFC-0018-P2P-INVENTORY-PROTOCOL (Nebula Network)

## 1. Arquitetura
A camada de sincronização inteligente (Inventory Protocol) da Nebula Network foi desenhada para otimizar o tráfego P2P. Em vez de injetar blocos e transações inteiras através de broadcast direto para todos (Gossip pesado), os nós anunciam resumos criptográficos (`ObjectHash`) através de pacotes de inventário. Essa arquitetura desacoplada e localizada em `pc_node/p2p` utiliza um design *Pull-based* garantindo que nenhum nó baixe novamente um objeto que já reside no seu banco local ou no cache transitório.

## 2. Inventory Manager
O `InventoryManager` atua como o ponto focal.
- Suas interfaces primárias são atiradas pelo `MessageRouter`.
- Coordena duas estruturas principais: um cache rápido de validação (`InventoryCache`) e uma fila assimétrica de aquisições pendentes (`InventoryQueue`).
- Ele empacota resumos (hashes) via método genérico `Announce()` que propaga o sumário à rede.
- Lida com eventos complexos, como timeout de objetos requisitados que nunca chegaram, mantendo registro na memória segura de `pendingReqs`.

## 3. Fluxo de Sincronização Inteligente
1. O Nó-A propaga um `MsgInventory` anunciando `Hash-X`.
2. O Nó-B recebe o anúncio no `ReceiveInventory()`.
3. O Nó-B averigua no seu `InventoryCache`. Se o Hash-X já for conhecido (Cache Hit), o anúncio é sumariamente **ignorado** (Zero tráfego extra).
4. Se desconhecido, o Nó-B encarrega o `InventoryQueue` de programar a busca.
5. O Worker do Queue dispara o `MsgGetData` (Pull Request) apontando para o Nó-A.
6. O Nó-A (através do Ledger real - aqui mockado) atira o `MsgData` de volta, contendo os bytes reais do Bloco/Transação.
7. O Nó-B, ao constatar o recebimento, oficializa a aquisição adicionando o hash ao Cache local e notificando o Gossip ou Storage para processamento via callbacks (`OnObjectDelivered`).

## 4. Cache (InventoryCache) e Deduplicação O(1)
O coração da contenção de banda é o cache local `map[string]InventoryCacheEntry`.
Sua checagem (`Contains`) é executada sob O(1) e englobada pelo `sync.RWMutex`.
Ao invés de estourar a RAM com milhões de chaves, um `Cleanup()` assíncrono rotineiramente limpa chaves cujo `TTL` expirou (CacheExpiration).

## 5. Fila Assíncrona (InventoryQueue)
Requisições Pull são despachadas paralelamente via canais go (`chan InventoryJob`).
- **MaxConcurrentRequests**: Limitado pelo tamanho de *MaxWorkers* instanciados no Init.
- Sem *Busy Wait*, o canal dorme esperando novos objetos para buscar.

## 6. Backpressure, Retry e Timeout
Para lidar com incertezas da internet:
- **QueueOverflow**: Se as pendências de Pull request ultrapassarem `maxSize`, a fila recusa novos pacotes, resguardando a RAM do nó.
- **Retry Limitado**: Um request de Pull (GetData) falhando em ir para o socket escala de volta pra fila (Backoff) até esgotar `maxRetries`.
- **Timeouts**: Se a rede transmitir o `GetData`, mas a entidade não responder com `MsgData` dentro de `reqTimeout`, a promessa é abortada, a pendência deletada e o callback `OnObjectTimeout` acorda módulos superiores da perda temporal.

## 7. Eventos Desacoplados
Toda etapa cruza com `InventoryEvents`: recebimentos de inventário, ignições de pull requests, detecção de objetos vazios, falhas de queue, timeouts e descobrimento. A malha orientada a eventos permite que a Blockchain se plugue futuramente sem violar a abstração P2P.

## 8. Estatísticas Thread-Safe
O `InventoryStatistics` gera dashboards ao vivo documentando eficiências e perdas: Hit Ratio do Cache (medindo quantos dados repetidos a rede tentou nos empurrar), Latência de Request (`AverageRequestTime`), contagem de timeouts e overflow de workers atuando.

## 9. Concorrência Segura e Zero Deadlock
Todo ponto de mutação na memória (Métricas, Cache, Dicionário de Responses) é resguardado por `sync.RWMutex` (ou Mutex puro no map das promessas `pendingReqs`). O sistema é isento de travamentos iterativos (Busy Waiting), suportando picos de tráfego agressivo sendo limitados exclusivamente por I/O.

## 10. Testes e Validação
A infraestrutura sobreviveu ao *Stress Test* desenhado para sobrecarga:
- Anúncios massivos distribuídos paralelamente via goroutines;
- Concorrência de recebimento simulando uma tempestade de inventários paralelos;
- Zero Panics ou vazamentos de goroutines validados através de context cancelations;
- Aprovado integralmente no detector interno de *Data Race* do Go Compiler.

## 11. Limitações Conhecidas
- A topologia atua baseada numa premissa de que todo *Announce* vem de alguém que de fato detém os dados. Nós maliciosos podem causar um ataque de negação de serviço anunciando hashes falsos provocando tráfego fantasma.
- O mapeamento temporário das requisições abertas em `pendingReqs` pode inflar se `reqTimeout` for estupidamente grande.

## 12. Próximas Expansões
- Implementação de um módulo de punição (Score System) conectado ao `SecurityManager` para penalizar (via Ban) nós que atiram falsos anúncios que resultam em *MsgNotFound* sucessivamente.
