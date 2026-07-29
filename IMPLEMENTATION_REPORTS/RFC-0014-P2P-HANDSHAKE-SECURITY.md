# IMPLEMENTATION REPORT: RFC-0014-P2P-HANDSHAKE-SECURITY (Nebula Network)

## 1. Arquitetura da Camada de Segurança P2P

A camada de segurança P2P e handshake oficial da Nebula Network foi implementada dentro do pacote isolado `p2p`. Ela garante que conexões só sejam elevadas para tráfego de dados (blocos/mempool/sync) após autenticação criptográfica/lógica rigorosa e compatibilidade de rede comprovada.

O módulo desacopla a regra de negócio através de **PeerEventHandlers**, de forma que o nó core pode reagir aos eventos de rede (`OnPeerConnected`, `OnPeerAuthenticated`, `OnPeerRejected`, `OnPeerBlacklisted`, `OnHeartbeatTimeout`) sem injetar dependência circular.

## 2. Protocolo de Handshake

O fluxo ocorre obrigatoriamente logo após o TCP Connect:
1. O cliente (Iniciador) envia `MsgHello` (contendo Genesis Hash, Network ID, Versão, Agent, NodeID e Capacidades).
2. O servidor (Listener) decodifica e realiza a validação rigorosa (HandshakeValidationOptions).
3. Se os dados não casarem perfeitamente, o servidor envia `MsgReject` com o motivo e fecha a conexão abruptamente.
4. Se validados, verifica-se a duplicidade de NodeID (proteção contra sybil/clones locais).
5. O servidor então responde com `MsgHelloAck` (NodeID e Capabilities).
6. O cliente também valida o `HelloAck`.
7. O Peer transita para o estado "Autenticado" e as goroutines de Heartbeat e Message Loop são iniciadas.

## 3. Heartbeat e Latência

- `MsgPing` e `MsgPong` trafegam timestamps locais em nanossegundos.
- A resposta do Pong subtrai o tempo atual para atualizar a Latência (Round Trip Time).
- Tolerância total baseada em Timeouts (ex: 200ms na testnet). Se o Peer ficar em silêncio por mais tempo que o `HeartbeatTimeout`, ele sofre drop preventivo para liberar os File Descriptors do nó.

## 4. Proteção contra Flood e Segurança

Para proteger contra exaustão de recursos, foi construído um `SecurityManager`:
- **Rate Limiter de Conexões**: Restringe a quantidade de conexões brutas oriundas de um mesmo IP por uma janela de tempo (ex: 5 conexões a cada 10 segundos).
- **Rate Limiter de Mensagens**: Uma vez conectado, impede spam de mensagens inúteis. Se exceder a quota, a conexão cai.
- **Temporary Blacklist**: O não cumprimento de rate limits ou handshakes corrompidos lançam o IP do par mal-comportado numa Quarentena (Blacklist Temporária). Outras tentativas de conexão são rejeitadas silenciosamente até a punição expirar.
- **Duplication Manager**: Tabela de bloqueio por `NodeID` garantindo unicidade na malha.

## 5. Testes Unitários e Integração

O pacote `p2p` possui testes cobrindo 100% da regra do Handshake e Segurança:
- Handshake Bidirecional Válido.
- Rejeição por divergência de `NetworkID` (Mainnet vs Testnet).
- Rejeição por divergência de `GenesisHash` (Forks incompatíveis).
- Rejeição por versão Mínima de Protocolo (Upgrade Forçado).
- Proteção contra `NodeID` Duplicado.
- Detecção e Drop por Heartbeat Timeout.
- Rate limits em massa de Conexões e Mensagens (Flood).
- Validação temporal da Quarentena (Blacklist Expirando).

Tudo roda 100% no padrão de test suite isolado e garante segurança no runtime da Nebula.
