# 10 IMPROVEMENT OPPORTUNITIES

### 1. Database ACID
- **Descrição:** Substituir Array em Memória e ioutil.WriteFile por key-value store (ex: LevelDB).
- **Motivação:** Garantir persistência sem data-loss; remover serialização JSON repetitiva inteira de O(N); viabilizar índices rápidos sem travar RAM.
- **Riscos:** Necessidade de conversão de dados pré-existentes.
- **Arquivos:** `ledger.go`
- **Complexidade:** Alta
- **Impacto esperado:** Escalabilidade Infinita de I/O, Redução de RAM para MB.
- **Prioridade:** 1 (Blocker)

### 2. Lock-free State Manager (Actor Model)
- **Descrição:** Remover `sync.Mutex` e gerenciar a Chain através de canais (Channels/Select).
- **Motivação:** Mutex retém requests TCP globais, causando timeout e partição da rede mesh sob estresse.
- **Riscos:** Eventuais data races em consultas síncronas requerem cuidados ao desenhar RPC interno.
- **Arquivos:** `ledger.go`, `network.go`
- **Complexidade:** Alta
- **Impacto esperado:** High Throughput, Não bloqueia a porta 5556 sob ataque ou carga.
- **Prioridade:** 2 (High)

### 3. Orphan Block & Fork Resolution
- **Descrição:** Modificar `handleNewBlock` para não rejeitar sumariamente blocos com `Index` ou `PreviousHash` alternativos. Guardar na memória e aplicar regra de Longest Chain.
- **Motivação:** A ausência de Fork Resolution dividirá fatalmente a rede na primeira discordância entre nós fisicamente distantes, inutilizando a moeda na WAN.
- **Riscos:** Vetor para spam se limite de ram-orphan não for fixado.
- **Arquivos:** `ledger.go`
- **Complexidade:** Muito Alta
- **Impacto esperado:** Resiliência de consenso P2P de nível profissional.
- **Prioridade:** 3 (High)

### 4. Cache UTXO / Account Balances
- **Descrição:** Em vez de fazer O(N) da Gênesis até Block[Last] para calcular saldo da Tx, manter um mapa em disco de `address -> balance`.
- **Motivação:** Envio de transação em redes avançadas levará horas apenas para verificar saldos se O(N) se mantiver.
- **Riscos:** Possível descompasso cache-storage se não atualizado atomicamente junto com bloco.
- **Arquivos:** `ledger.go`
- **Complexidade:** Média
- **Impacto esperado:** Verificação instantânea de transação O(1).
- **Prioridade:** 4 (Medium)

### 5. Safe Init (Truncate Protection)
- **Descrição:** Mudar `saveLedger` para escrever em `ledger.json.tmp` e renomear atomicamente (`os.Rename`), preservando a leitura.
- **Motivação:** Solução de low-hanging fruit que impede que um shutdown destrua 100% dos dados.
- **Riscos:** Nenhum.
- **Arquivos:** `ledger.go`
- **Complexidade:** Baixa
- **Impacto esperado:** Sobrevivência a quedas de luz.
- **Prioridade:** 1 (Quick Win)
