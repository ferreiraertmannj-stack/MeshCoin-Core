# 07 WALLET

## Criação e Chaves
As carteiras são criadas usando a curva elíptica `secp256k1`. A geração de endereços e gestão do par de chaves parece estar suportada tanto em código Go quanto no app Flutter (implementação Dart). No backend Go, usa-se a biblioteca `decred/dcrd/dcrec/secp256k1`.

## Assinatura
As transações assinam o pacote (Timestamp + Amount + Fee + Sender + Receiver) com a chave privada. No arquivo Go, o código (implícito em `VerifyTransaction` que não foi lido por completo) verifica a assinatura ECDSA do `SenderPubKey`. Há também o campo `PQCSignature` previsto.

## Saldo e Histórico
Calculado iterando sobre todo o `ledger.json` e somando entradas e subtraindo saídas. Não existe modelo UTXO, é um sistema baseado em Contas/Saldos calculados on-the-fly.

## Fonte da Verdade
A única fonte da verdade é a reconstrução do estado baseada na lista sequencial de blocos confirmados no `ledger.json`.
