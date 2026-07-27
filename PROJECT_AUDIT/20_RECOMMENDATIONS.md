# 20 RECOMMENDATIONS

Nenhuma destas recomendações deve ser executada sem planejamento futuro.

1. **CRÍTICO:** Trocar imediatamente o modelo de dados de arquivo `ledger.json` inteiro por um banco Key-Value DB, gravando blocos por chaves binárias.
2. **CRÍTICO:** Implementar salvamento atômico (`file.tmp` -> rename `file.json`) na função `saveLedger()` antes da troca de DB.
3. **CRÍTICO:** Adicionar rate limits severos no socket TCP e checagem de assinatura *antes* de deserializar pacotes de 100MB de transações forjadas.
4. **ALTO:** Proteger o Cloud Upload. Um nó da Cloud deve requerer prova de posse de chave privada válida com saldo, ou PoW Hashcash antes de alocar 100MB de seu disco.
5. **MÉDIO:** Refatorar pacote `main` do `pc_node` em submódulos isolados: `network`, `core`, `storage`.
