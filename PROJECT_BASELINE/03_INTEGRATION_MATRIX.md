# 03 INTEGRATION MATRIX

## Mapeamento de Dependências

- **Mobile App (Flutter) -> PC Node (Go):** Integração Crítica. O mobile atua apenas como assinador/interface e depende cegamente da saúde do PC Node para espalhar tx e baixar blocos da LAN.
- **PC Node -> Nebula Cloud:** Integração Moderada. O PC Node fará upload a cada 10 blocos. Se a Cloud cair, o Node deve arquivar temporariamente e não crashar.
- **Python Scripts -> PC Node / Cloud:** Experimental. Ferramentas externas de auditoria criptográfica. Sem dependência direta.

## Integrações Críticas Identificadas
1. `Mempool <-> Consensus`: A exclusão de transações da Mempool (Mempool Mutex) mediante a validação de um novo bloco (Ledger Mutex) cria um risco cruzado de deadlock.
2. `TCP Listener <-> Ledger Disk IO`: A recepção de pacotes congela a rede se o IO local do ledger for sobrecarregado, expondo a integração entre Networking e Storage como falha de design.
