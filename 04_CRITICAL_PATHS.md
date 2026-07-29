# 04 CRITICAL PATHS

## 1. Persistência do Ledger
- `handleNewBlock` -> Mutex Lock -> Validation -> Array Append -> `json.Marshal` -> Disk Write -> Mutex Unlock.
*(Caminho extremamente custoso e perigoso).*

## 2. Mineração e Broadcast
- Calculo NeonHash -> Sucesso -> Propaga `NEW_BLOCK` -> Outros nós validam -> Rejeitam se conflito / Aceitam e gravam.

## 3. Criação de Carteira
- (App Mobile) Gera Entropia -> Secp256k1 Key Pair -> ECDSA Sign -> Propaga Address -> Escuta Blockchain por balance.

## 4. Transferência
- Sign Tx -> TCP `NEW_TRANSACTION` -> Validador Mempool -> Mineração -> Ledger.

## 5. Entrada de Novo Nó
- Inicializa vazio -> UDP 5555 Broadcast -> Aceita novas tx/blocos TCP -> *(Não há Fetch Histórico = Nó inútil)*.

## 6. Sincronização
- Atualmente manual/passiva.
