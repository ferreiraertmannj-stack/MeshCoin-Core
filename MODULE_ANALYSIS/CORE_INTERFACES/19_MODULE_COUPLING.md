# 19 MODULE COUPLING

O acoplamento é extremamente ALTO entre:
- `Network` e `Ledger` (Variável global `ledger` e `PendingTransactions` é acessada por pacotes da rede).
O acoplamento é BAIXO entre:
- `Crypto` (puro) e os demais componentes.
