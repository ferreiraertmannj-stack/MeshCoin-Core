# 05 RELEASE STRATEGY

## 1. Alpha
- **Critério:** Código atual estabilizado, sem perdas catastróficas.
- **Escopo:** JSON db corrigido para Atomic Save, Mempool protegida e Cloud requer auth básica. Nós funcionam sem corromper estado em queda de energia.

## 2. Developer Preview
- **Critério:** Troca do núcleo concluída (LevelDB integrado, sem Global Locks).
- **Escopo:** Desenvolvedores conseguem iniciar nós que sincronizam milhares de blocos em poucos segundos em teste local. APIs de depuração expostas.

## 3. Beta
- **Critério:** Camada de Rede Mesh implementada e comissionada (Kademlia/DHT funcional).
- **Escopo:** Dispositivos móveis encontram Full Nodes em redes diferentes sem IPs fixos (com auxílio de peers de bootstrap).

## 4. Release Candidate (RC)
- **Critério:** Auditorias de segurança completas e zero regressão de rede.
- **Escopo:** Teste de Stress contínuo; rede aguenta spam de conexões sem paralisar consenso. Consenso resolve forks perfeitamente (Longest chain robusta).

## 5. Stable (Mainnet)
- **Critério:** Gênesis Block gerado oficialmente; nós móveis compilados e prontos para distribuição em massa; Cloud distribuindo fragmentos entre pares reais com incentivos ativados.
