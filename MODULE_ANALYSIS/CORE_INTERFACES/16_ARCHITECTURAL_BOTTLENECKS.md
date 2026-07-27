# 16 ARCHITECTURAL BOTTLENECKS

## Gargalos de Comunicação
1. **Network -> Ledger:** O socket aguarda o Lock Global (`sync.RWMutex`) do Ledger, transformando redes assíncronas em processos síncronos e lentos.
2. **Ledger -> Storage:** Dump integral do array JSON paralisa todo o IO.
