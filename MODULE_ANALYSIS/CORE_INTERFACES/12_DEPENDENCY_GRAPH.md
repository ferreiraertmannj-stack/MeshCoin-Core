# 12 DEPENDENCY GRAPH

## Grafo Lógico de Dependências
```text
[API/Cloud] -----> [Node Daemon]
                      |
[Wallet (Flutter)] -> [Node Network (TCP)]
                      |
[Mesh Discovery] ---> [Network] ---> [Consensus (Verify)]
                                        |
                                        V
[Storage (Disk)] <------------------ [Ledger]
                                        |
[Mining (App)] ----> [Consensus] -------+
```
