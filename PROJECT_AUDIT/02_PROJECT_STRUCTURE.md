# 02 PROJECT STRUCTURE

## Estrutura de Diretórios
- `/docs/architecture/`: Documentação oficial imutável.
- `/meshcoin_flutter/`: Aplicativo mobile principal, UI, cliente P2P mobile. Depende da rede local para propagar transações.
- `/pc_node/`: Código Go do Full Node. Responsável pelo consenso, manutenção do `ledger.json` e mineração via TCP/UDP. Depende da Nebula Cloud para long-term storage.
- `/nebula_cloud/`: Servidor de armazenamento descentralizado (Go) que recebe fragmentos de dados (Reed-Solomon).
- `/tools/`: Reservado para ferramentas internas de engenharia.
- `/*.py` (raiz): Scripts Python de prototipação (p.ex. `chat_mesh.py`, `rede_p2p.py`, `pqc_crypto.py`).

## Dependências entre Componentes
- **Flutter App** -> **PC Node**: O app mobile transmite transações P2P via UDP/TCP, que chegam ao PC Node para validação.
- **PC Node** -> **Nebula Cloud**: O PC Node faz o upload do ledger e snapshots a cada X blocos (ex: 10 blocos) para a nuvem.
- **Python Scripts** -> **N/A**: Operam de maneira independente, principalmente como simuladores e validadores de conceito para a arquitetura P2P e PQC.
