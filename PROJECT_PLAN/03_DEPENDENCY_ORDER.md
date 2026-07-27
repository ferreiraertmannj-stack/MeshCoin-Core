# 03 DEPENDENCY ORDER

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
