# Relatório Estratégico e Técnico: Nébula Cloud

**Data:** Abril de 2026  
**Autor:** Manus AI  
**Cliente:** Jean Ertmann (Ertmann Tech)  
**Projeto:** Nébula Cloud (Armazenamento Descentralizado Brasileiro)

---

## 1. Visão Geral do Projeto

O projeto **Nébula Cloud** propõe uma infraestrutura de armazenamento em nuvem descentralizada, sustentável e altamente segura, utilizando hardware reaproveitado (como smartphones flagship e SSDs) e operando sob a arquitetura Edge Computing [1]. O modelo de negócios é baseado em afiliados (Nébula Box Pro), com promessa de redução de custos em 25% frente às grandes provedoras (AWS, Azure, Google Cloud) e divisão de receitas 50/50 [2].

A análise do MVP atual revelou uma base conceitual forte, mas com falhas críticas na implementação técnica que impediam o funcionamento básico (distribuição e recuperação de arquivos).

---

## 2. Diagnóstico Técnico e Correções do MVP

Durante a auditoria do código, identificamos e corrigimos os seguintes problemas que impediam o MVP de operar corretamente:

### 2.1. Falha Crítica na Criptografia e Fragmentação (`fragmenter.py`)
*   **Problema:** O algoritmo AES-256-GCM exigia uma chave de exatamente 32 bytes. A implementação original usava um método inseguro (`ljust`) para preencher a senha, o que causava erros de validação e perda de integridade na reconstrução (erro `InvalidTag`). Além disso, o tamanho dos fragmentos não considerava padding adequado para divisão perfeita.
*   **Correção:** Implementamos derivação de chave criptográfica robusta usando **PBKDF2HMAC** com SHA-256 e um *salt* fixo. Ajustamos a lógica de divisão de fragmentos com *padding* dinâmico para garantir que todos os fragmentos Reed-Solomon tenham exatamente o mesmo tamanho, permitindo a reconstrução perfeita do arquivo.

### 2.2. Erro de Integração na Recuperação (`tracker.py`)
*   **Problema:** O script de recuperação tentava ler o arquivo `meta.json` do diretório temporário de recuperação, mas esse arquivo nunca era baixado dos nós (apenas os `.bin` eram recuperados). Isso resultava em `FileNotFoundError`.
*   **Correção:** Modificamos o `tracker.py` para extrair os metadados diretamente do arquivo de mapa (`.map.json`) salvo localmente e recriar o `meta.json` na pasta de recuperação antes de chamar a função de reconstrução.

### 2.3. Vulnerabilidades no Node Daemon (`node_daemon.py`)
*   **Problema:** O servidor FastAPI não possuía tratamento de erros, validação de nomes de arquivos (vulnerabilidade de *Path Traversal*) ou logging adequado. Além disso, faltava a dependência `python-multipart` no ambiente para lidar com upload de arquivos.
*   **Correção:** Adicionamos validação rigorosa de rotas, tratamento de exceções com retornos HTTP adequados (400, 404, 500), e um sistema de logging completo. O daemon agora também reporta o tamanho total armazenado no endpoint de `/status`.

### 2.4. Melhorias no Heartbeat (`heartbeat.py`)
*   **Problema:** O monitoramento era básico e não mantinha histórico persistente de instabilidades.
*   **Correção:** Adicionamos logging estruturado, cálculo de capacidade total em GB, tratamento de timeouts de rede e salvamento de um histórico rotativo (`nos_historico.json`) para auditoria futura de SLA.

---

## 3. Análise do Modelo de Negócios e Infraestrutura

O Masterplan da Fase 1 [2] apresenta uma proposta de valor agressiva e atrativa:

| Métrica | Mercado Tradicional | Nébula Cloud | Vantagem |
| :--- | :--- | :--- | :--- |
| **Armazenamento (GB/mês)** | ~R$ 0,11 | R$ 0,08 | 27% mais barato |
| **Egress / Download (GB)** | ~R$ 0,50 | R$ 0,37 | 26% mais barato |
| **Custo de Hardware (Nó 8TB)** | N/A (Datacenter) | R$ 7.000,00 | CapEx Descentralizado |
| **ROI Mensal Estimado** | N/A | 3,30% a 6,27% | Superior ao CDI |

### Pontos Fortes do Modelo:
1.  **Soberania de Dados:** O apelo de manter os dados fragmentados e processados localmente (Edge) é um forte argumento de vendas para prefeituras e órgãos públicos [1].
2.  **Sustentabilidade (ESG):** O reaproveitamento de hardware de alta performance (S22 Ultra) e arquitetura ARM reduz drasticamente o consumo energético [2].
3.  **Modelo de CapEx:** Transferir o custo do hardware para o afiliado permite escalar a rede rapidamente sem comprometer o fluxo de caixa da Ertmann Tech.

### Riscos e Gargalos:
1.  **Disponibilidade de Banda:** O afiliado arca com a internet local. Conexões residenciais geralmente possuem *upload* assimétrico (menor que o download), o que pode gargalar a recuperação de dados pelos clientes.
2.  **SLA e Uptime:** Datacenters garantem 99.9% de uptime. Smartphones em residências estão sujeitos a quedas de energia, resets acidentais e falhas de roteador. A redundância Reed-Solomon (10+4) permite a perda de até 4 nós simultâneos, mas em uma rede pequena (ex: 13 nós iniciais), isso representa um risco moderado.

---

## 4. Roadmap para Lucratividade e Escala

Para transformar o MVP corrigido em um produto comercialmente viável e lucrativo, recomendamos o seguinte plano de ação:

### Fase 1: Estabilização e Showroom (Meses 1-3)
*   **Ação Técnica:** Implementar autenticação mútua (mTLS ou tokens JWT) entre o Tracker e os Node Daemons. Atualmente, qualquer um na rede pode deletar ou enviar fragmentos para a porta 8000.
*   **Ação Comercial:** Montar o showroom com 3 nós em Extrema-MG, conforme o plano original [2]. Utilizar esses nós para hospedar dados não-críticos de parceiros locais a custo zero, gerando *cases* de sucesso e métricas reais de *uptime*.

### Fase 2: Captação de Afiliados e Resiliência (Meses 4-6)
*   **Ação Técnica:** Desenvolver um painel web (Dashboard) para os afiliados acompanharem seus ganhos, uso de disco e status do nó em tempo real.
*   **Ação Comercial:** Vender os primeiros 10 "Nébula Box Pro" para investidores locais. O foco deve ser a transparência do ROI de ~3.3% ao mês [1].

### Fase 3: Contratos Institucionais (Meses 7-12)
*   **Ação Técnica:** Implementar integração compatível com a API do Amazon S3. Para que prefeituras e empresas usem a Nébula Cloud, elas precisam conseguir plugar seus sistemas de backup (ex: Veeam, Bacula) sem reescrever código. Um *gateway* S3 é obrigatório.
*   **Ação Comercial:** Atacar licitações públicas com o argumento ESG ("Nuvem Verde") e Soberania Digital. O foco deve ser backup frio (Cold Storage) e arquivamento de longo prazo, onde a latência é menos crítica.

---

## 5. Conclusão

O código do MVP da Nébula Cloud foi **totalmente corrigido** e agora é capaz de fragmentar, distribuir, criptografar e recuperar arquivos com sucesso. O modelo de negócios é inovador e possui forte apelo comercial, especialmente no nicho de B2G (Business to Government) focado em soberania de dados.

O próximo grande desafio técnico será garantir a segurança da comunicação entre os nós e desenvolver uma interface compatível com o padrão S3 da indústria, o que destravará a adoção por clientes corporativos.

---
**Referências:**
[1] Ertmann Tech. *NÉBULA CLOUD: MASTERPLAN DE INFRAESTRUTURA ELITE (80TB – FASE 1)*. Abril 2026.
[2] Ertmann Tech. *NÉBULA CLOUD*. Abril 2026.
