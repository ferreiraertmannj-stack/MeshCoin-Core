# 04 RISK MATRIX

| Problema | Probabilidade | Impacto | Prioridade | Risco de Regressão | Risco Arquitetural |
|---|---|---|---|---|---|
| **REF-001 (Storage JSON)** | 100% | Fatal (OOM/IO) | 1 | Alto (Requer converter todo o parseamento) | Crítico (Muda como os dados existem) |
| **REF-002 (Global Locks)** | 100% | Severo (Travamento) | 2 | Médio (Pode introduzir data races) | Alto (Muda modelo de concorrência) |
| **REF-003 (Data Loss)** | 50% | Fatal (Amnésia) | 3 | Baixo | Baixo |
| **REF-004 (Cloud DDoS)** | 90% | Severo (Disk Full) | 1 | Baixo (Restringe acesso) | Médio (Exige auth header) |
| **REF-005 (Falso Mesh)** | 100% | Mod. (Só LAN) | 4 | Alto (Muda totalmente os sockets) | Crítico (Troca protocolo de rede) |
