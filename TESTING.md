# Testing and Validation Infrastructure

Este documento descreve como executar a infraestrutura de validação do projeto Nebula Network. Nenhuma lógica simulada (mocks) está presente na base de testes; os testes dependem da estabilização da arquitetura, e falharão intencionalmente ou exibirão alertas onde refatorações (ex: LevelDB, Canais Go) são pré-requisitos.

## Estrutura
- `test/unit/`: Testes de unidade (isolados por módulo).
- `test/integration_test/`: Testes E2E cruzando processos.
- `tools/validation/`: Ferramentas Python para validação estática e arquitetural.

## Scripts Automatizados
Use os scripts PowerShell (`.ps1`) no Windows ou Shell (`.sh`) no Linux/Mac:
- `run_all_tests`: Executa testes unitários e de integração.
- `run_health_check`: Consolida status geral.
- `run_architecture_validation`: Verifica violações arquiteturais estruturais.
- `run_invariant_validation`: Checa garantias de sistema contínuas.

## Continuous Integration (CI)
A pipeline no GitHub Actions (`.github/workflows/ci.yml`) está configurada para buildar, testar e rodar todos os checkers arquiteturais a cada Push e PR.
