# 01 EXECUTIVE SUMMARY

## Resumo Executivo
O projeto Nebula Network consiste em um ecossistema offline-first focado em comunicação em malha (Mesh Networking) e um armazenamento distribuído (Nebula Cloud). O componente blockchain (MeshCoin Core) existe como uma camada de registro e recompensa para nós da rede.

## Estado Geral
O projeto encontra-se em um estado funcional de protótipo (Alpha), possuindo implementações híbridas em Go (Full Node, Cloud) e Python (Protótipos), juntamente com um aplicativo Flutter para dispositivos móveis. Há forte ênfase em criptografia (transição planejada para PQC) e resiliência offline.

## Arquitetura Encontrada
- **Mobile/Client:** App Flutter que lida com UI, Carteira, e comunicação Bluetooth/Wi-Fi Direct.
- **Node PC:** Implementação em Go (`pc_node/main.go`) que lida com validação de blocos TCP/UDP, memória de transações (mempool), e interface com o armazenamento Cloud.
- **Cloud Node:** Servidor Go (`nebula_cloud/node_daemon.go`) para armazenamento descentralizado (fragmentação via Reed-Solomon).
- **Scripts em Python:** Protótipos de roteamento, tracker, mineração e integração criptográfica.

## Riscos
- Múltiplas linguagens para lógicas chave (Python para prototipação e Go para Node) podem causar inconsistências no protocolo.
- O consenso baseado em arquivo local JSON (`ledger.json`) apresenta alto risco de dessincronização e conflitos em um ambiente distribuído se a camada de rede não garantir consistência forte.
- Dificuldade estática na mineração sem um mecanismo robusto de ajuste contínuo (aparentemente dependente de halving hardcoded).

## Pontos Fortes
- Foco explícito em resistência à censura e operação off-grid.
- Inovação no algoritmo de mineração focado em CPU/Mobile (NeonHash).
- Divisão clara de responsabilidades entre cliente móvel, full node de validação e infraestrutura de cloud.

## Pontos Fracos
- Arquitetura de armazenamento do ledger em arquivo único JSON.
- Forte acoplamento de lógica de negócio em scripts Python dispersos.
- Gerenciamento de rede híbrido (UDP/TCP) propenso a particionamento de rede.
