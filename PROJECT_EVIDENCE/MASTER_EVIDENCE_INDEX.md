# MASTER EVIDENCE INDEX

## Quantidade Total de Problemas Analisados
5 evidências críticas estruturadas (representando as anomalias centrais do ecossistema mapeado).

## Quantidade por Severidade
- **Crítico:** 2 (Ledger Atomicity, Network DDoS)
- **Alto:** 2 (Global Lock, JSON Storage)
- **Médio:** 1 (False Mesh)
- **Baixo:** 0

## Arquivos Mais Críticos (Focos de Instabilidade)
1. `pc_node/ledger.go` (Responsável por 3 dos 5 problemas)
2. `nebula_cloud/node_daemon.go` (Responsável pela brecha de Cloud)
3. `pc_node/network.go` (Responsável pelo falso-mesh e gargalo de I/O)

## Módulos Mais Afetados
- **pc_node** (A fundação do consenso e validação demonstra arquitetura prematura inadequada para ambientes hostis e distribuídos).

## Ordem Sugerida de Correção
A fila de correção (Prioridade Técnica) deve seguir a fundação do sistema (do persistente para o roteamento):
1. `001_LEDGER_ATOMICITY.md` (Risco de perda total)
2. `002_NETWORK_DDOS.md` (Risco de exaustão externa)
3. `005_JSON_STORAGE.md` (Permitirá a correção arquitetural de locks e desempenho)
4. `003_GLOBAL_LOCK.md` (Sem JSON bloqueante, os locks podem ser fine-grained)
5. `004_FALSE_MESH.md` (Pode aguardar libp2p ou implementação DHT robusta)

## Dependências entre problemas
A correção do `003_GLOBAL_LOCK.md` (Performance/Gargalos) está intimamente acoplada à solução de `005_JSON_STORAGE.md`. Substituir o JSON por um banco KV (LevelDB) resolverá naturalmente os tempos de bloqueio de I/O, exigindo que a refatoração do storage preceda a refatoração de concorrência.
