import os

out_dir = "PROJECT_PLAN"
os.makedirs(out_dir, exist_ok=True)

files = {
    "01_MASTER_REFACTORING_PLAN.md": """# 01 MASTER REFACTORING PLAN

## 1. ID: REF-001 | Nome: Migração de Storage para Key-Value DB (Substituir JSON)
- **Severidade:** Crítico
- **Arquivos envolvidos:** `pc_node/ledger.go`
- **Dependências:** Nenhuma.
- **Complexidade:** Alta
- **Estimativa relativa:** G
- **Ordem correta:** 1
- **Justificativa:** JSON marshal de toda a blockchain inviabiliza performance e escalabilidade e causa lock contínuo. Nenhuma refatoração de throughput pode ocorrer antes de corrigir o gargalo de disco.

## 2. ID: REF-002 | Nome: Refatoração de Concorrência (Remover Global Locks)
- **Severidade:** Alto
- **Arquivos envolvidos:** `pc_node/ledger.go`, `pc_node/network.go`
- **Dependências:** REF-001
- **Complexidade:** Média
- **Estimativa relativa:** M
- **Ordem correta:** 2
- **Justificativa:** Após ter I/O rápido (KV DB), podemos introduzir lock granular ou Actor Model (canais) para permitir paralelismo nas conexões TCP.

## 3. ID: REF-003 | Nome: Prevenção de Data Loss (Salvamento Atômico / Safe Init)
- **Severidade:** Crítico
- **Arquivos envolvidos:** `pc_node/ledger.go`
- **Dependências:** REF-001 (Mas mitigável imediatamente se ignorado o plano ideal). Assumimos dependência de REF-001.
- **Complexidade:** Baixa
- **Estimativa relativa:** P
- **Ordem correta:** 3
- **Justificativa:** Tratar a falha de leitura (corrupção) em vez de simplesmente sobrescrever com um bloco Gênesis novo.

## 4. ID: REF-004 | Nome: Autenticação na Nebula Cloud
- **Severidade:** Crítico
- **Arquivos envolvidos:** `nebula_cloud/node_daemon.go`
- **Dependências:** Nenhuma.
- **Complexidade:** Média
- **Estimativa relativa:** M
- **Ordem correta:** 4
- **Justificativa:** Impedir DDoS e enchimento imediato do HD de nós da rede na nuvem exigindo assinatura ou Hashcash.

## 5. ID: REF-005 | Nome: Implementação de Roteamento Mesh Real (B.A.T.M.A.N. / DHT)
- **Severidade:** Médio
- **Arquivos envolvidos:** `pc_node/network.go`
- **Dependências:** REF-002
- **Complexidade:** Muito Alta
- **Estimativa relativa:** GG
- **Ordem correta:** 5
- **Justificativa:** Atualmente é um broadcast cego de rede local TCP/UDP. Requer sistema distribuído real (libp2p ou equivalente) para operar globalmente.
""",

    "02_EXECUTION_ROADMAP.md": """# 02 EXECUTION ROADMAP

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
""",

    "03_DEPENDENCY_ORDER.md": """# 03 DEPENDENCY ORDER

Grafo Lógico e Ordem Estrita de Desbloqueio:

```text
[REF-001: Key-Value DB]
    │
    ├─> [REF-002: Remover Global Locks] (Desbloqueado pois IOPS não gargala mais a goroutine)
    │       │
    │       └─> [REF-005: Roteamento Mesh Real] (Desbloqueado pois a rede pode lidar com throughput massivo)
    │
    └─> [REF-003: Safe Init e Prevenção de Data Loss] (Se implementado diretamente no novo Storage driver)

[REF-004: Autenticação Cloud]
    │
    └─> (Correção independente, pode correr em paralelo no Sprint 3 ou 1, pois foca em nebula_cloud e não pc_node)
```
""",

    "04_RISK_MATRIX.md": """# 04 RISK MATRIX

| Problema | Probabilidade | Impacto | Prioridade | Risco de Regressão | Risco Arquitetural |
|---|---|---|---|---|---|
| **REF-001 (Storage JSON)** | 100% | Fatal (OOM/IO) | 1 | Alto (Requer converter todo o parseamento) | Crítico (Muda como os dados existem) |
| **REF-002 (Global Locks)** | 100% | Severo (Travamento) | 2 | Médio (Pode introduzir data races) | Alto (Muda modelo de concorrência) |
| **REF-003 (Data Loss)** | 50% | Fatal (Amnésia) | 3 | Baixo | Baixo |
| **REF-004 (Cloud DDoS)** | 90% | Severo (Disk Full) | 1 | Baixo (Restringe acesso) | Médio (Exige auth header) |
| **REF-005 (Falso Mesh)** | 100% | Mod. (Só LAN) | 4 | Alto (Muda totalmente os sockets) | Crítico (Troca protocolo de rede) |
""",

    "05_RELEASE_STRATEGY.md": """# 05 RELEASE STRATEGY

## 1. Alpha
- **Critério:** Código atual estabilizado, sem perdas catastróficas.
- **Escopo:** JSON db corrigido para Atomic Save, Mempool protegida e Cloud requer auth básica. Nós funcionam sem corromper estado em queda de energia.

## 2. Developer Preview
- **Critério:** Troca do núcleo concluída (LevelDB integrado, sem Global Locks).
- **Escopo:** Desenvolvedores conseguem iniciar nós que sincronizam milhares de blocos em poucos segundos em teste local. APIs de depuração expostas.

## 3. Beta
- **Critério:** Camada de Rede Mesh implementada e comissionada (Kademlia/DHT funcional).
- **Escopo:** Dispositivos móveis encontram Full Nodes em redes diferentes sem IPs fixos (com auxílio de peers de bootstrap).

## 4. Release Candidate (RC)
- **Critério:** Auditorias de segurança completas e zero regressão de rede.
- **Escopo:** Teste de Stress contínuo; rede aguenta spam de conexões sem paralisar consenso. Consenso resolve forks perfeitamente (Longest chain robusta).

## 5. Stable (Mainnet)
- **Critério:** Gênesis Block gerado oficialmente; nós móveis compilados e prontos para distribuição em massa; Cloud distribuindo fragmentos entre pares reais com incentivos ativados.
"""
}

for filename, content in files.items():
    with open(os.path.join(out_dir, filename), "w", encoding="utf-8") as f:
        f.write(content)
