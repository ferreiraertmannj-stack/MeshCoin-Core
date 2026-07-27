# 04 MODULE ANALYSIS

## PC Node (Go)
- **Responsabilidade**: Manter estado global da blockchain, minerar blocos, propagar estado pela rede P2P (TCP/UDP).
- **Maturidade**: Alpha.
- **Riscos**: Serialização do ledger inteiro em JSON, sem uso de Merkle Trees eficientes para sync parcial.

## Nebula Cloud (Go)
- **Responsabilidade**: Armazenamento fragmentado e redundante.
- **Maturidade**: Proof of Concept.
- **Riscos**: Falta de incentivo direto automatizado e verificação de disponibilidade (Proofs of Spacetime/Replication).

## Mobile Client (Flutter)
- **Responsabilidade**: Interação do usuário, carteira (chaves secp256k1), chat.
- **Maturidade**: Beta/Funcional.
- **Riscos**: Vazamento de chaves privadas no storage do celular se não for utilizado Keystore/Secure Enclave nativo.

## Python Prototypes
- **Responsabilidade**: Simulação PQC, Reed-Solomon, Tracking.
- **Maturidade**: Experimental.
