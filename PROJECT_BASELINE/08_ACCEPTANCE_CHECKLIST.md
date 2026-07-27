# 08 ACCEPTANCE CHECKLIST

Qualquer PR ou Fix proposto no Nebula Network deverá cumprir:

- [ ] 1. Não utiliza Mocks no pacote produtivo. Fakes apenas restritos aos arquivos `_test.go` ou diretório `/test`.
- [ ] 2. Não apaga a documentação, apenas adiciona ou altera status.
- [ ] 3. Se altera estrutura TCP/UDP, o mobile Flutter é compatível.
- [ ] 4. Testes unitários para nova lógica cobrem cenários negativos explícitos.
- [ ] 5. Concorrência: Código submetido não bloqueia o IO principal com `sync.Mutex` ao longo de chamadas de disco (`ioutil.Write`).
- [ ] 6. Nenhuma refatoração altera o Bloco Gênesis ou endereços legados.
- [ ] 7. Invariantes Mantidos (Sistema Descentralizado preservado, Off-grid possível).
