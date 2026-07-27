# 08 ATOMICITY

### Escrita em Disco
- **Não existe atomicidade real**.
- O Go `ioutil.WriteFile()` abre, trunca (O_TRUNC) e escreve. Se a energia cair entre o `Open()` e o fim da gravação, o `ledger.json` restará corrompido (tamanho 0 ou parcial).

### Adição na Memória
- **Garantida pelo Mutex.** O array em Go nunca sofrerá *data race* graças ao `ledger.mu`. Portanto, transições *em RAM* são atômicas. Mas, se a RAM corromper ou for perdida, o disco não sustenta as ACID properties devido ao I/O ser inseguro.
