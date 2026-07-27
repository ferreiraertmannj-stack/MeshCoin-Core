# 09 RISK ACCEPTANCE

## Riscos Aceitáveis
1. **Lentidão em Dispositivos Low-End (Mobile):** A mineração NeonHash consumirá muita bateria; este risco é tolerado temporariamente até o balanceamento definitivo do ASIC resistance.
2. **Desconexão de Peers em Redes Celulares (4G/5G Nat via CGNAT):** Até o DHT robusto estar funcional, dependemos de IPs públicos do PC e aceita-se temporariamente instabilidade mesh.
3. **Inconsistência da Mempool Visual:** O App Flutter pode apresentar transações como "Aguardando" por longos períodos se a malha isolar o celular. Tolerável.

## Riscos Inaceitáveis
1. **Perda de Ledger Local:** Corrupção de arquivos de estado mestre nunca pode destruir o trabalho validado.
2. **Execução Remota (RCE) via Pacotes P2P:** Ausência de limites TCP/BufferOverflow.
3. **Locks Globais Permanentes:** Desligar a rede inteira por IO engasgado descaracteriza a "alta disponibilidade" da Nebula.
