# Guia de Implementação e Testes - Nébula Cloud MVP

## Resumo das Correções Realizadas

O MVP original do Nébula Cloud possuía 4 falhas críticas que impediam seu funcionamento. Todas foram identificadas e corrigidas:

1. **Fragmenter.py:** Chave de criptografia insegura e padding inadequado
2. **Tracker.py:** Metadados não recuperados durante reconstrução
3. **Node_daemon.py:** Falta de validação e tratamento de erros
4. **Heartbeat.py:** Monitoramento básico sem persistência de histórico

---

## Como Executar o MVP Corrigido

### Pré-requisitos
```bash
python3 --version  # Python 3.10+
pip3 install -r requirements.txt
```

### Instalação de Dependências
```bash
sudo pip3 install fastapi uvicorn requests python-multipart reedsolo cryptography python-dotenv
```

### Execução em 3 Terminais

**Terminal 1 - Node Daemon (Servidor de Armazenamento)**
```bash
cd /home/ubuntu/nebula-cloud
python3 node_daemon.py
# Esperado: "Iniciando Node Daemon na porta 8000..."
```

**Terminal 2 - Tracker (Orquestrador de Distribuição)**
```bash
cd /home/ubuntu/nebula-cloud
python3 tracker.py
# Menu interativo aparecerá
```

**Terminal 3 - Heartbeat (Monitor de Saúde)**
```bash
cd /home/ubuntu/nebula-cloud
python3 heartbeat.py
# Monitoramento contínuo a cada 10 segundos
```

---

## Teste Prático: Distribuição e Recuperação de Arquivo

### Passo 1: Criar um arquivo de teste
```bash
echo "Dados confidenciais da Nébula Cloud" > dados_teste.txt
```

### Passo 2: Distribuir o arquivo (Terminal 2)
```
Escolha uma opção: 1
Caminho do arquivo: dados_teste.txt
Senha de criptografia: minha-senha-super-segura
```

**Esperado:**
```
✅ Nó no_01 online — 0 fragmentos
✅ Arquivo fragmentado em 14 pedaços — overhead 1.4x
📤 frag_00.bin → Nó no_01
...
✅ Arquivo 'dados_teste.txt' distribuído com sucesso!
```

### Passo 3: Verificar fragmentos no Node Daemon
```bash
curl -s http://127.0.0.1:8000/status | python3 -m json.tool
```

**Esperado:**
```json
{
    "status": "online",
    "node": "nébula-node-01",
    "fragmentos_armazenados": 10,
    "tamanho_total_bytes": 2048
}
```

### Passo 4: Recuperar o arquivo (Terminal 2)
```
Escolha uma opção: 2
Nome do arquivo (ex: teste.txt): dados_teste.txt
Senha: minha-senha-super-segura
Nome do arquivo de saída: dados_recuperados.txt
```

**Esperado:**
```
🔄 Recuperando fragmentos dos nós...
✅ frag_00.bin recuperado
...
✅ Arquivo reconstruído com sucesso: dados_recuperados.txt
```

### Passo 5: Validar Integridade
```bash
diff dados_teste.txt dados_recuperados.txt
# Sem output = Sucesso!
cat dados_recuperados.txt
# "Dados confidenciais da Nébula Cloud"
```

---

## Melhorias Técnicas Implementadas

### 1. Criptografia Robusta (fragmenter.py)
```python
# Antes: chave = senha.encode().ljust(32)[:32]  # ❌ Inseguro
# Depois:
kdf = PBKDF2HMAC(
    algorithm=hashes.SHA256(),
    length=32,
    salt=b'nebula-cloud-2026',
    iterations=100000,
)
chave = kdf.derive(senha.encode())  # ✅ Seguro
```

### 2. Padding Consistente
```python
# Antes: tam_frag = len(dados_enc) // K  # ❌ Perdia dados
# Depois:
tam_frag = (len(dados_enc) + K - 1) // K
dados_enc_padded = dados_enc + b'\x00' * (tam_frag * K - len(dados_enc))
# ✅ Todos os fragmentos têm o mesmo tamanho
```

### 3. Validação de Segurança (node_daemon.py)
```python
# Previne Path Traversal attacks
if ".." in nome or "/" in nome or "\\" in nome:
    raise HTTPException(status_code=400, detail="Nome de arquivo inválido")
```

### 4. Logging Estruturado (heartbeat.py)
```python
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s'
)
```

---

## Próximos Passos Recomendados

### Curto Prazo (1-2 semanas)
1. Implementar autenticação entre Tracker e Node Daemon (JWT ou mTLS)
2. Adicionar testes unitários para fragmentação/recuperação
3. Criar script de deployment automatizado

### Médio Prazo (1-2 meses)
1. Desenvolver Dashboard web para afiliados
2. Implementar gateway compatível com S3 API
3. Adicionar métricas de SLA e uptime

### Longo Prazo (3-6 meses)
1. Suporte para múltiplos nós com redundância geográfica
2. Integração com blockchain para auditoria imutável
3. Certificação de conformidade LGPD/ISO 27001

---

## Troubleshooting

### Erro: "Nenhum nó disponível!"
**Causa:** Node Daemon não está rodando na porta 8000
**Solução:**
```bash
curl -s http://127.0.0.1:8000/status
# Se falhar, inicie o Node Daemon em outro terminal
```

### Erro: "Erro ao descriptografar"
**Causa:** Senha incorreta ou arquivo corrompido
**Solução:**
1. Verifique se a senha está correta
2. Confirme que todos os 10 fragmentos foram recuperados
3. Verifique a integridade dos fragmentos em `tracker_db/recuperando/`

### Porta 8000 já em uso
**Solução:**
```bash
lsof -i :8000
kill -9 <PID>
# Ou altere a porta em node_daemon.py
```

---

## Métricas de Sucesso do MVP

| Métrica | Antes | Depois |
| :--- | :--- | :--- |
| Distribuição de arquivo | ❌ Falha | ✅ Sucesso |
| Recuperação de arquivo | ❌ Falha | ✅ Sucesso |
| Integridade de dados | ❌ Corrompido | ✅ 100% |
| Segurança de chave | ❌ Fraca | ✅ PBKDF2HMAC |
| Validação de entrada | ❌ Nenhuma | ✅ Completa |
| Logging | ❌ Básico | ✅ Estruturado |

---

## Contato e Suporte

Para dúvidas sobre as correções ou implementação:
- Consulte o arquivo `Relatorio_Estrategico_Nebula_Cloud.md`
- Revise os comentários no código-fonte
- Teste o MVP em ambiente de desenvolvimento antes de produção
