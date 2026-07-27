# 01 MASTER REFACTORING PLAN

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
