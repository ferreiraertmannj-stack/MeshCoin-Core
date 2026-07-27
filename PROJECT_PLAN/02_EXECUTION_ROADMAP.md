# 02 EXECUTION ROADMAP

## Sprint 1: Fundação de Dados e Consistência
- **Objetivo:** Garantir que o estado da blockchain seja indestrutível e escalável no disco.
- **Atividades:** Migrar `ledger.json` para LevelDB/BadgerDB, implementar testes de integridade e recarregamento seguro (Safe Init/Atomic Save).

## Sprint 2: Alta Disponibilidade e Concorrência (Core)
- **Objetivo:** Permitir que o nó suporte milhares de conexões e blocos concorrentes.
- **Atividades:** Remover `sync.RWMutex` globais do `ledger.go` e da Mempool, reescrevendo para modelo de canais Go (Actor model) ou transações ACID de banco.

## Sprint 3: Proteção de Perímetro e Cloud
- **Objetivo:** Eliminar vetores críticos de ataque DDoS e DoS na infraestrutura de nuvem.
- **Atividades:** Implementar autenticação via assinatura na API da Nebula Cloud, verificação de TTL em pacotes TCP, e limites de conexão por IP no nó validador.

## Sprint 4: Arquitetura de Rede Mesh P2P
- **Objetivo:** Descentralizar o roteamento e permitir descoberta global sem centralizadores (além de LAN).
- **Atividades:** Substituir broadcast UDP cego por DHT (Distributed Hash Table) via libp2p ou implementação customizada inspirada em Kademlia. Integrar routing ad-hoc (B.A.T.M.A.N. layer).
