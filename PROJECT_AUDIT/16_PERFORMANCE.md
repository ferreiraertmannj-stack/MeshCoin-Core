# 16 PERFORMANCE

## Gargalos Identificados
- `saveLedger()` (Go): Usa JSON `MarshalIndent` carregando todo o histórico na RAM para salvar no disco a cada bloco confirmado.
- Tratamento de conexões TCP em `handleConnection` gerando uma nova goroutine (desempenho normal em Go, mas perigoso com floods P2P devido à ausência de socket limits).
- Broadcasts UDP agressivos (cada 5 seg) em rede local, impactante para smartphones (`battery drain`).
- Vetor do `NeonHash` em memória gasta ~4KB. Na escala de múltiplas threads do minerador, é extremamente amigável ao cache L1 da CPU, mas a verificação síncrona bloqueia a validação.
