# IMPLEMENTATION REPORT: RFC-0015-PEER-DISCOVERY (Nebula Network)

## 1. Arquitetura da Descoberta de Peers
O ecossistema de descoberta de peers da Nebula Network foi implementado no pacote `pc_node/p2p`, mantendo o isolamento absoluto da lógica de negócio (Consenso, Blockchain, Fast Sync, etc). A arquitetura é centrada no `DiscoveryManager`, que orquestra a fila assíncrona de descoberta (`PeerDiscoveryQueue`), o banco de pontuações persistentes (`PeerStore`), e o fallback de infraestrutura (`SeedManager`).

## 2. Mensagens do Protocolo
Quatro mensagens oficiais foram definidas para gerenciar a topologia:
- `MsgGetPeers`: Solicita a um peer conhecido uma lista limitada de seus peers.
- `MsgPeers`: Resposta contendo um array de `PeerRecord` (NodeID, IP, Port, Versão, Capacidades).
- `MsgPeerAnnouncement`: Mensagem enviada espontaneamente para anunciar a entrada do nó na vizinhança.
- `MsgPeerGoodbye`: Mensagem de graceful shutdown informando a saída do nó.

## 3. Fluxo de Descoberta Assíncrono e Limites
Após o Handshake de sucesso, o nó solicita automaticamente `MsgGetPeers`.
Quando um pacote de `MsgPeers` chega:
1. Sofre validação de schema (IP válido, Porta > 0).
2. Passa pelo `SecurityManager` (Blacklist check e Rate Limit de conexões por IP).
3. Entra na fila `PeerDiscoveryQueue`, que executa Deduplicação O(1) usando Hash Map para evitar congestionamento na tentativa de discagem de clones num mesmo ciclo.
4. Workers assíncronos retiram peers da fila e emulam a discagem (Dial). Se bem-sucedido, invocam o evento `OnPeerConnected`.

## 4. Peer Scoring e Persistência
Para proteger a rede e otimizar rotas, a `PeerRecord` possui métricas contínuas:
- **Reliability (Confiabilidade)**: Sobe a cada conexão bem-sucedida, desce a cada falha de resposta.
- **Failures**: Se ultrapassar o limite fixado (ex: 5 falhas consecutivas), o peer é ejetado.
- **Latency / Uptime / LastSuccess**: Mantêm o registro de saúde atualizado.
**Persistência**: Ao invés de modificar a interface base do `storage.Engine` (o que violaria a diretiva de não alteração), a engine foi projetada para utilizar arquivos locais serializados, salvando e restaurando o estado total de *Scores* após reinícios (`peers.json`).

## 5. Seed Nodes e Bootstrap Automático
O `SeedManager` contém endereços hardcoded confiáveis. Eles *nunca* são usados em operação normal, ativando-se apenas sob três condições críticas:
1. O nó liga pela primeira vez (tabela de peers vazia).
2. Isolamento total da rede (todos os peers na tabela morreram).
3. Perda completa ou corrupção do banco local de peers.

## 6. Random Walk Discovery
Além de pedir peers na conexão inicial, existe uma goroutine de varredura ativa. O `Random Walk Loop` acorda periodicamente (ex: a cada 2 minutos), seleciona um peer pseudo-aleatório da tabela persistida, e "espalha a rede" perguntando sobre novos nós, garantindo capilaridade orgânica da malha sem estressar as rotas.

## 7. Limpeza Automática (Garbage Collection)
Um processo em background (`cleanupLoop`) acorda a cada 10 minutos para:
- Limpar a cache de deduplicação da fila.
- Remover peers que estão sem se comunicar com sucesso há mais de 24 horas.
- Ejetar permanentemente peers cujos IPs tenham caído na Blacklist do `SecurityManager` recentemente.

## 8. Eventos Desacoplados
Toda operação reage disparando os _PeerEventHandlers_ (`OnPeerDiscovered`, `OnPeerValidated`, `OnPeerConnected`, `OnPeerRemoved`, `OnPeerExpired`, `OnDiscoveryFinished`). A camada superior de rede os mapeia em ações concretas sem gerar dependência circular.

## 9. Limitações Conhecidas
- A rede não possui suporte DHT maduro; depende estritamente do Random Walk entre clusters conectados.
- A persistência local em JSON atende cenários com alguns milhares de nós. Redes super densas (>1M nós) exigirão acoplamento LevelDB/Badger nativo se o footprint de memória se tornar restrito.
