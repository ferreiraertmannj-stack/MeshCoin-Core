# 17 INTERFACE RISKS

1. **Ausência de Contratos Fortes (Interfaces Go):** Módulos conversam via struct literal e globais. Não há `type LedgerStore interface` o que impede injeção de dependência e testes unitários.
2. **API Cloud sem Middleware Auth:** O limite da interface Cloud não tem bloqueio criptográfico; aceita form-data cego.
