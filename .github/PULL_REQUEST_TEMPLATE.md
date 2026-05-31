<!--
Obrigado pelo PR! Para acelerar a review, preencha as seções abaixo.
Veja CONTRIBUTING.md antes de submeter: https://github.com/finance-sh/finance-sh/blob/main/CONTRIBUTING.md
-->

## Tipo de mudança

<!-- Marque com x o que se aplica. -->

- [ ] 🐛 `fix` — corrige bug
- [ ] ✨ `feat` — nova funcionalidade
- [ ] ♻️ `refactor` — refactor sem mudança de comportamento externo
- [ ] ⚡ `perf` — performance
- [ ] 📝 `docs` — documentação
- [ ] ✅ `test` — testes
- [ ] 🔧 `chore` — build, deps, infra
- [ ] 👷 `ci` — CI/CD
- [ ] 🌐 `i18n` — tradução

## Descrição

<!-- O que muda? Por quê? Contexto que ajude o revisor. -->

## Issue relacionada

<!-- "Closes #123" fecha automaticamente. "Refs #123" apenas referencia. -->

Closes #

## Como testei

<!-- Descreva os passos manuais e/ou automatizados que validaram a mudança. -->

- [ ] `go test ./... -race -count=1` passa (backend)
- [ ] `npm run test` passa (frontend / landing)
- [ ] `npm run build` passa sem erro de TS
- [ ] Testei manualmente o fluxo afetado
- [ ] Adicionei/atualizei testes cobrindo a mudança
- [ ] Adicionei/atualizei migrations (up + down) se mexeu em schema

## Screenshots / GIFs

<!-- Para mudanças visuais. Antes/depois ajuda muito. -->

## Checklist do contribuidor

- [ ] Commits seguem [Conventional Commits](https://www.conventionalcommits.org/) e estão `Signed-off-by:` (`git commit -s`)
- [ ] Código formatado: `gofmt -s` (backend) + Prettier (frontend/landing)
- [ ] Lint passa: `golangci-lint run ./...` (backend) + `npm run lint` (frontend/landing)
- [ ] Sem `// nolint` ou `// @ts-ignore` sem justificativa em comentário
- [ ] Strings visíveis ao usuário estão em `frontend/src/i18n/locales/{pt-BR,en,es}/translation.json`
- [ ] Money manipulado como `int64` em centavos (nunca `float64`)
- [ ] PII nova vai criptografada via `crypto.EncryptedString` (ou justificada nos comentários)
- [ ] Audit log captura a alteração quando aplicável
- [ ] Documentação atualizada (README/docs/CHANGELOG) se mudou comportamento público
- [ ] Sem breaking change OU breaking change documentado no footer do commit (`BREAKING CHANGE: ...`)

## Considerações de segurança / LGPD

<!--
Marque com x se aplicável e descreva.
- Mudou superficie de auth/autorização?
- Mudou tratamento de PII?
- Mudou retenção/exclusão de dados?
- Adicionou nova dependência externa?
-->

- [ ] Não toca auth/autorização/criptografia/PII
- [ ] Toca, e descrevi mitigações abaixo:

## Para mantenedores

<!-- Não preencher. Espaço pra revisor anotar. -->

- [ ] Squash & merge
- [ ] Atualizar CHANGELOG
- [ ] Rotular release (`semver:patch`/`minor`/`major`)
