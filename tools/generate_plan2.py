import os

out_dir = "PROJECT_PLAN"
os.makedirs(out_dir, exist_ok=True)

files = {
    "06_MILESTONES.md": """# 06 MILESTONES

- **M1: Ledger Resiliente:** Storage trocado para DB nativo e dados persistem perfeitamente a panics/crashes.
- **M2: Lock-Free Consensus:** Nó suporta 5.000 requisições TCP/sec sem corrupção da mempool ou chain-halt.
- **M3: Nebula Secured:** Cloud Node rejeita payloads arbitrários e verifica assinaturas do Core.
- **M4: Mesh Foundation:** DHT ativada; rede descobre peers via WAN sem IPs fixos injetados.
- **M5: Fork Resolution:** Algoritmo de Longest Chain incorporado. Rede resolve conflitos sozinha.
- **M6: Alpha Release:** Deploy e testes de carga em rede viva distribuída.
""",

    "07_SPRINT_BACKLOG.md": """# 07 SPRINT BACKLOG

## Tarefa: Migração para LevelDB
- **Descrição:** Reescrever `saveLedger` e a estrutura base do array `Chain` para leitura/gravação indexada via goleveldb.
- **Dependências:** Nenhuma.
- **Estimativa:** 13 pontos (Alta)
- **Criticidade:** Crítica
- **Critério de aceite:** O nó inicializa lendo do DB. O salvamento não sobrecarrega a RAM, testes passam.

## Tarefa: Atomic Fallback temporário
- **Descrição:** Enquanto M1 não termina, alterar `ioutil.WriteFile` para gravar em `tmp` e dar rename, protegendo JSON.
- **Dependências:** Nenhuma.
- **Estimativa:** 2 pontos (Baixa)
- **Criticidade:** Crítica
- **Critério de aceite:** Ao simular queda de energia no save, o arquivo anterior não é corrompido.

## Tarefa: Autenticação na API Cloud
- **Descrição:** Modificar handlers HTTP do `node_daemon.go` para validar assinatura ECDSA.
- **Dependências:** Assinatura na Wallet já funcional (Dart e Go).
- **Estimativa:** 5 pontos (Média)
- **Criticidade:** Crítica
- **Critério de aceite:** HTTP 401 para uploads forjados ou maiores que 100MB; HTTP 200 para uploads corretos.

## Tarefa: Remover Mutex Global
- **Descrição:** Utilizar Channels em `pc_node/ledger.go` para enfileirar requests de `NEW_BLOCK` num único state-manager thread.
- **Dependências:** Tarefa N1 LevelDB concluída (para não criar fila infinita bloqueada por IO lento).
- **Estimativa:** 8 pontos (Alta)
- **Criticidade:** Alta
- **Critério de aceite:** Testes de stress de blocos concorrentes viajam na fila sem causar Data Race.

## Tarefa: Roteamento Mesh via Kademlia
- **Descrição:** Integrar ou construir abstração DHT em substituição ao Broadcast de UDP e map de TCP fixo em `network.go`.
- **Dependências:** Mutex Removido.
- **Estimativa:** 21 pontos (GG)
- **Criticidade:** Alta
- **Critério de aceite:** Nó PC em SP descobre Nó PC na Alemanha via DHT bootstrap.
""",

    "08_SUCCESS_CRITERIA.md": """# 08 SUCCESS CRITERIA

## Blockchain Pronta
Quando suportar no mínimo 1.000 TPS, for capaz de salvar o estado instantaneamente no disco, resolver forks organicamente descartando órfãos e reincorporando cadeias mais pesadas.

## Wallet Pronta
Quando chaves privadas forem protegidas nativamente por Secure Enclave (mobile), e as assinaturas puderem validar sem requerer nós centralizadores.

## Consensus Pronto
Quando puder bloquear injeção massiva de nós maliciosos (Sybil), calcular NeonHash rapidamente e não depender de locks sequenciais lentos.

## Mesh Pronto
Quando a comunicação entre dispositivos puder saltar de A -> B -> C transparentemente sem internet (via mobile) ou usar DHT puro sobre WAN.

## Nebula Pronta
Quando armazenar fragmentos, garantir disponibilidade, autenticar remetentes e recompensá-los on-chain.

## Projeto Pronto para Mainnet
Quando testes de chaos-engineering demonstrarem resiliência contra as falhas registradas, os milestones arquiteturais atingidos, e código inteiramente auditado por terceiros.
""",

    "09_TEST_STRATEGY.md": """# 09 TEST STRATEGY

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
""",

    "10_DEFINITION_OF_DONE.md": """# 10 DEFINITION OF DONE

Uma tarefa de engenharia na Nebula Network SÓ está concluída quando:

- [ ] **Código Funcionando:** A feature opera de ponta a ponta sem erros não tratados em logs.
- [ ] **Arquitetura Preservada:** Os contratos de API, serialização e assinaturas base respeitam os arquivos estáticos na pasta `/docs/architecture`.
- [ ] **Invariantes Preservados:** As regras de negócio vitais do documento `SYSTEM_INVARIANTS.md` não foram comprometidas (ex: segurança de rede, descentralização real, etc).
- [ ] **Testes Passando:** Existem testes automatizados provando o caminho feliz e os erros de limite da implementação (Cobertura).
- [ ] **Documentação Atualizada:** Se afetou o design, os documentos em `PROJECT_PLAN/` ou `docs/` receberam um patch referenciando a mudança.
- [ ] **Sem Regressão:** O módulo atual não quebrou os outros (ex: mudar o bloco no Go e esquecer do parse no Flutter).
- [ ] **Sem Mocks:** Código "Mock", "Stub" ou "Simulado" não são mesclados em commits finais (conforme diretriz de FASE 0 e FASE 1).
"""
}

for filename, content in files.items():
    with open(os.path.join(out_dir, filename), "w", encoding="utf-8") as f:
        f.write(content)
