# 01 CURRENT SYSTEM STATE

## Estado Atual por Módulo

- **Blockchain (Ledger):** Parcialmente Funcional
  - *Justificativa:* Consegue minerar blocos e salvar histórico JSON. Mas a resiliência contra arquivos corrompidos está quebrada e não há mecanismo nativo forte de reorg de cadeia (Forks).
- **Consensus:** Experimental
  - *Justificativa:* O algoritmo NeonHash é executado e validado corretamente, mas o nó carece de verificação em larga escala e resolve empates de blocos pela primeira vez que são vistos, sem `Longest Chain` claro implementado adequadamente no TCP stream atual.
- **Wallet:** Funcional
  - *Justificativa:* A geração de chaves `secp256k1` e verificação de assinatura de pacotes de dados operam de forma condizente e determinística.
- **Mesh:** Não Implementado (Falso) / Parcialmente Funcional (LAN apenas)
  - *Justificativa:* O protocolo fala em "Mesh" e roteamento multi-hop offline, mas o que existe no `network.go` é estritamente um TCP broadcast e um UDP de LAN. Sem roteadores Ad-Hoc ou DHT funcionais.
- **Mining:** Funcional
  - *Justificativa:* Mineração NeonHash opera perfeitamente com controle de dificuldade estático e verificação P2P coerente.
- **Networking:** Parcialmente Funcional
  - *Justificativa:* A base TCP roda estavelmente para 1:1, mas está propensa a loop infinito de rotas devido à ausência de pacotes roteáveis indexados e é frágil a DoS (lock global).
- **Nebula Cloud:** Experimental
  - *Justificativa:* Endpoint exposto, consegue enviar blocos, mas não existe autenticação, criptografia nativa no pacote base sem uso dos scripts paralelos ou prova de storage amarrada à wallet via smart-contract.
- **Storage (Local):** Quebrado (Em Escala)
  - *Justificativa:* Lidar com dump de `ledger.json` inteiro não permite escalabilidade além de algumas dezenas de megabytes antes de esgotar CPU/RAM de um nó pequeno.
- **Synchronization:** Parcialmente Funcional
  - *Justificativa:* Nodes pegam pacotes de broadcasts ao estarem online (listening). Recuperar um node atrasado desde o Gênesis é inviável no código atual.
- **Crypto:** Funcional
  - *Justificativa:* ECDSA funciona, e o pacote PQC (Dilithium) tem base sólida, restando apenas substituição final dos stubs.
