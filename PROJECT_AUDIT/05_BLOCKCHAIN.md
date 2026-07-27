# 05 BLOCKCHAIN

## Como Funciona
A blockchain é mantida em memória e salva em disco como `ledger.json` no nó PC. Os blocos são empilhados sequencialmente. O bloco Gênesis é hardcoded.

## Quem Cria e Valida
- **Criação**: Nós PC e Mobile criam blocos através de mineração.
- **Validação**: Feita no momento da recepção em `handleNewBlock` (no arquivo `ledger.go`), validando índice, hash prévio, e PoW.

## Armazenamento
Arquivos JSON serializados locais (`ledger.json`) no PC. Nuvem Nebula Cloud armazena backups a cada 10 blocos.

## Dificuldade e Hash
Utiliza o algoritmo `NeonHash` (vetor em memória de 4KB, operações matemáticas pseudo-vetoriais iteradas 128 vezes, finalizando com SHA-256).

## Resolução de Conflitos
A implementação atual de `ledger.go` **rejeita sumariamente** qualquer bloco cujo índice seja menor ou igual ao último conhecido, ou cujo PreviousHash não bata. Falta um mecanismo robusto de reconciliação de forks (Longest Chain real com reorg).

## Bugs e Riscos
- Risco de split brain irreversível na rede sem um processo automático de rollback de blocos órfãos.
- Leitura do JSON inteiro na memória na inicialização.
