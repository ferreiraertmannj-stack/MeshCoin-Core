# 10 DEFINITION OF DONE

Uma tarefa de engenharia na Nebula Network SÓ está concluída quando:

- [ ] **Código Funcionando:** A feature opera de ponta a ponta sem erros não tratados em logs.
- [ ] **Arquitetura Preservada:** Os contratos de API, serialização e assinaturas base respeitam os arquivos estáticos na pasta `/docs/architecture`.
- [ ] **Invariantes Preservados:** As regras de negócio vitais do documento `SYSTEM_INVARIANTS.md` não foram comprometidas (ex: segurança de rede, descentralização real, etc).
- [ ] **Testes Passando:** Existem testes automatizados provando o caminho feliz e os erros de limite da implementação (Cobertura).
- [ ] **Documentação Atualizada:** Se afetou o design, os documentos em `PROJECT_PLAN/` ou `docs/` receberam um patch referenciando a mudança.
- [ ] **Sem Regressão:** O módulo atual não quebrou os outros (ex: mudar o bloco no Go e esquecer do parse no Flutter).
- [ ] **Sem Mocks:** Código "Mock", "Stub" ou "Simulado" não são mesclados em commits finais (conforme diretriz de FASE 0 e FASE 1).
