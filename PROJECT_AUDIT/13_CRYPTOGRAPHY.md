# 13 CRYPTOGRAPHY

## Implementações Reais
Usa de fato `secp256k1` (Go) para assinaturas clássicas e `AES-256-GCM` na fragmentação (verificado scripts Python / Go na nuvem). 

## Implementações Simuladas
O código contém placeholders estruturais para assinaturas Pós-Quânticas (campo `PQCSignature` vazio, mas scripts PQC Python com referências ao Kyber/Dilithium apontam para integração real em andamento). O NeonHash foi validado como real e customizado.

## Pontos Críticos
O vetor em memória inicializado no `NeonHash` depende unicamente do hash SHA-256 do record do bloco. Sem um pool dinâmico extra de estado global (como em Ethash), esse algoritmo ainda é vulnerável a precomputação agressiva, embora mais resistente a ASICs simples.
