# 09 TEST STRATEGY

## Unitários
- **Foco:** Funções puras (NeonHash, ECDSA Verify, Halving logic).
- **Meta:** 80% de cobertura nos algoritmos centrais isolados.

## Integração
- **Foco:** Pipeline TCP/UDP em conjunto com Ledger Storage.
- **Testes:** Forjar nós maliciosos mandando transações inválidas.

## Stress & Carga
- **Foco:** Disparar milhares de transações e centenas de blocos UDP simultâneos na porta 5556 para garantir que o Actor Model / DB lidam sem Memory Leak.

## Recuperação (Chaos)
- **Foco:** Matar o processo (kill -9) do PC node durante gravação de bloco e garantir que o nó boote sem perder a cadeia válida até o penúltimo bloco.

## Segurança
- **Foco:** Pen-test na porta 8000 (Cloud) e fuzzing em portas 5555/5556 (Node).

## Forks e Sincronização
- **Foco:** Inicializar dois nós de origens desincronizadas e forçar o menor adotar a chain maior (reorg) corretamente, verificando o saldo da wallet resultante.
