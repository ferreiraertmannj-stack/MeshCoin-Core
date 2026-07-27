# 19 TECHNICAL DEBT

## Arquitetural
- Armazenamento em arquivos planos JSON, exigindo refatoração imediata para LevelDB.
- Sistema P2P sem algoritmo real Kademlia/DHT para descoberta em WAN, dependendo de Broadcasts 255.255.255.255 (só LAN).

## Código
- Dependência de variáveis globais e Mutexes únicos (`ledger.mu`, `mempoolMutex`) em vez de canais (channels) ou DB locks em Go.

## Documentação
- A documentação afirma roteamento mesh B.A.T.M.A.N. inteligente, mas o código oficial é TCP estático.
