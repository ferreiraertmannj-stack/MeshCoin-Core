# Estratégia de Testes - Nebula Network

Esta pasta (atualmente vazia de código) hospedará a futura matriz contínua de Integração e Testes (CI/CD) e scripts autônomos.

## Evolução Futura
1. Implementação de Testes Puros em Go (`go test`) para o núcleo criptográfico, Consensus e Ledger.
2. Automação P2P (Testnets em Containêres Docker) orquestrados aqui, criando 5 a 20 nós isolados em rede bridge para observar partições de rede.
3. Integração com o projeto Flutter (`flutter test`) via mocks locais apenas da porta TCP e testes End-to-End nativos automatizados.
4. Testes de Stress de Carga implementados provavelmente em Python, simulando envio de milhares de pacotes `NEW_TRANSACTION`.

A infraestrutura de teste *jamais modificará* o core sem garantia prévia contra as Máquinas de Estados mapeadas em `PROJECT_BASELINE/05_STATE_MACHINES.md`.
