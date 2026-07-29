# IMPLEMENTATION REPORT: RFC-0013-P2P-FASTSYNC-TRANSPORT (Nebula Network)

## 1. Arquitetura da Camada de Transporte Fast Sync

O transporte P2P para o Fast Sync da Nebula Network foi implementado utilizando comunicação TCP direta. A arquitetura desacopla a regra de negócio da mineração/blockchain da comunicação do nó durante a inicialização, garantindo sincronização hiper-rápida. 

Os componentes chave são:
- **`TransportEncoder` e `TransportDecoder`**: Camadas exclusivas de serialização, que envelopam JSON puro com uma estrutura base (Header) que previne ataques por saturação de Buffer. O Header possui 4 bytes contendo o tamanho do Payload, 1 byte para o tipo e 1 byte reservado para Compressão.
- **`TCPPeer`**: Implementação 100% funcional da interface `Peer`, conectada a Soquetes reais. Utiliza a camada de decodificação de transporte para bloquear reads/writes com Timeouts por requisição.
- **`MsgGetHeaders` e `MsgGetBlocks`**: Protocolos dedicados ao Fast Sync. A estrutura não sobrepõe os Pings/Pongs padrão do nó de mempool.

## 2. Timeouts, Heartbeat e Recuperação

Cada conexão P2P no `TCPPeer` possui deadlines de 5 a 10 segundos, forçando um drop (desconexão) se o par for muito lento para preencher o Buffer (slowloris mitigation).

**Recuperação**:
- O `DownloadWorker` notifica `PeerPool` através de `AddFailure()` quando um read timeout, error, ou bad header ocorre.
- No Pool, a cada request a `BestPeer()`, o algoritmo analisa se as falhas de um par ultrapassam o `MaxFailures = 3`. 
- Caso afirmativo, o score é ignorado e ele é removido em tempo-real (eviction list), desconectando os sockets persistentes do lado do cliente.

## 3. Limitações (Size Limits)

As seguintes constantes evitam uso excessivo de RAM (OOM):
- `MaxBlocksPerMessage = 500`: Nunca tentar mandar milhões de chunks na mesma via.
- `MaxPayloadSize = 10MB`: Rejeita silenciosamente dados mal-formados antes de decodificar e processar o JSON via memory heap.

## 4. Serialização e Compressão

O Payload foi arquitetado para suportar `Compression byte` nativamente, com `0x00` como "None" e espaços reservados (Gzip, Snappy, Zstd). Neste momento, os bytes correm brutos. A serialização interna continua com a flexibilidade do JSON via interface `json.Unmarshal(tm.Payload, v)`, porém, nenhum outro arquivo na aplicação toca no protocolo TCP, exceto o módulo de Transporte.

## 5. Testes Implementados e Resultados

Testes robustos foram colocados no pacote `peer_tcp_test.go` para:
- Spawnar mock server na porta efêmera.
- Realizar PING/PONG com round-trip timing validation.
- Troca simulada de Header Requests.
- Troca de Block Requests.
- Estouro de Buffer `InvalidLength` para testar MaxPayloadSize overflow protection.
