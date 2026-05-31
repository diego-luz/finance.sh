# Contribuindo para finance.sh

Obrigado por considerar contribuir. **finance.sh** é controle financeiro open-source self-hosted (AGPL-3.0) para pessoa física, MEI e microempresa no Brasil. Este guia cobre tudo que você precisa para abrir uma issue, mandar um patch ou traduzir.

> 🇺🇸 English version: see [CONTRIBUTING.en.md](CONTRIBUTING.en.md) (em breve).

---

## Sumário

- [Antes de começar](#antes-de-começar)
- [Como reportar bug](#como-reportar-bug)
- [Como pedir feature](#como-pedir-feature)
- [Setup de desenvolvimento](#setup-de-desenvolvimento)
- [Stack do projeto](#stack-do-projeto)
- [Padrões de código](#padrões-de-código)
- [Testes](#testes)
- [Mensagens de commit](#mensagens-de-commit)
- [Branches e PR](#branches-e-pr)
- [Traduções (i18n)](#traduções-i18n)
- [Licença e DCO](#licença-e-dco)
- [Comunidade](#comunidade)

---

## Antes de começar

- Leia o [README.md](README.md) e a [arquitetura](docs/ARCHITECTURE.md).
- Veja se sua ideia/bug **já está em uma [issue aberta](https://github.com/finance-sh/finance-sh/issues)**.
- Para mudanças grandes (refactor, nova área do produto, mudança de stack), **abra uma issue de discussão antes** de escrever código — economiza retrabalho.
- Tudo neste projeto está em pt-BR como idioma primário. Issues, PRs e código podem ser em pt-BR ou en — escolha a maioritária do thread.

---

## Como reportar bug

Abra issue com o template `Bug report` e inclua:

- **Versão** (`git rev-parse HEAD` ou tag) e plataforma (Docker / bare-metal Go).
- **Passos para reproduzir** — mínimos, idealmente com `curl` ou screenshot.
- **Esperado vs. observado.**
- **Logs:** `docker compose logs backend --tail=200`, console do navegador, e (se rodando bare-metal) saída do processo.
- **Ambiente:** SO, versão do Docker, versão do navegador, browser local OU host (LAN IP é gotcha conhecida — veja [docs/SECURITY.md](docs/SECURITY.md)).

Não cole secrets/tokens/JWT em issues públicas. Sanitize.

---

## Como pedir feature

Abra issue `Feature request` e responda:

1. **Problema** que a feature resolve (ponto de vista do usuário, não da solução).
2. **Quem é afetado** — PF, MEI, microempresa, dev self-host, etc.
3. **Alternativas** que você considerou.

O projeto é 100% open-source (AGPL-3.0) — não há edição paga. Features entram no roadmap conforme prioridade e fit.

---

## Setup de desenvolvimento

### Pré-requisitos

| Ferramenta | Versão mínima | Motivo |
|---|---|---|
| Docker + Docker Compose v2 | 24+ | Stack inteira sobe por compose |
| Go | 1.22 | Backend |
| Node.js | 20 LTS | Frontend + Landing |
| Make (opcional) | — | Atalhos no `Makefile` |
| golang-migrate (opcional) | v4 | Migrations manuais |

### Subir o stack

```bash
git clone https://github.com/finance-sh/finance-sh
cd finance-sh
cp .env.example .env          # defaults dev funcionam de cara
docker compose up -d --build
```

Acesse:

| Serviço | URL |
|---|---|
| App (SPA + API via Nginx) | http://localhost:8090 |
| Swagger (dev) | http://localhost:8090/swagger/index.html |
| Postgres (cliente local) | `127.0.0.1:5433` (use `docker exec finance-sh-postgres psql` ou DBeaver/TablePlus) |
| Landing (dev) | `cd landing && npm run dev` → http://localhost:5174 (deploy externo; fora do compose) |

Usuário demo (SEED=true): `demo@finance.sh` / `demo1234`. Super-admin: `super@finance.sh` / `super1234`. Troque ambos no primeiro login.

### Dev sem Docker (loop rápido backend)

```bash
docker compose up -d postgres                   # única dependência
cd backend
cp .env.example .env                            # ajustar DB_HOST=localhost, DB_PORT=5433
go run ./cmd/api                                # roda em :8090
```

### Dev sem Docker (loop rápido frontend)

```bash
cd frontend
npm install
npm run dev                                     # Vite em :5173, proxy automático pra /api
```

---

## Stack do projeto

| Camada | Tech |
|---|---|
| App (binário único) | Go 1.22 · chi · GORM · PostgreSQL 16 · SPA embutida (`go:embed`) + API + scheduler in-process |
| Frontend SPA | React 18 + TS + Vite 6 + Tailwind + React Query + Zustand + Recharts + lucide-react · PWA · i18n pt-BR/en/es |
| Anexos | Postgres BYTEA (sem object store externo) |
| TLS / edge | reverse proxy do operador (Traefik/Caddy/Nginx Proxy Manager) — não embutido |
| Crypto | AES-256-GCM field-level para PII (notes, 2FA secret) |
| Auth | JWT access+refresh + bcrypt + TOTP 2FA + lockout |

Stack default = **2 containers** (postgres + app). Arquitetura completa: [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md).

---

## Padrões de código

### Go (backend)

- Formatação: `gofmt -s` obrigatório. CI rejeita diferença.
- Lint: `golangci-lint run ./...`. Zero `// nolint` sem comentário explicando.
- Layout: Clean Architecture — handlers → services → repositories → entities. Sem cross-skip.
- Erros: `fmt.Errorf("X: %w", err)` para wrap, `errors.Is/As` para check. Sem `panic` em código de produção.
- Money: **sempre `int64` em centavos**. Nunca `float64` para valores monetários.
- Time: `time.Time` UTC nos modelos; conversão pra timezone só na borda (handler).
- Context: receber `ctx context.Context` em handlers, services e repos que façam I/O.
- Migrations: `backend/internal/database/migrations/NNNNNN_descricao.{up,down}.sql`. Numeração sequencial. Down obrigatório.

### TypeScript (frontend)

- Formatação: Prettier (`.prettierrc` do projeto).
- Lint: ESLint + `@typescript-eslint`. Sem `any` salvo justificativa em comentário.
- Componentes: function components + hooks. Sem class components.
- Estado: Zustand para global, React Query para server state, `useState`/`useReducer` para local.
- Forms: React Hook Form + Zod para validação. Mensagens de erro em pt-BR (use i18n).
- Estilo: Tailwind utility-first. Tokens em `tailwind.config.ts`. Sem CSS-in-JS.
- Acessibilidade: `aria-*` quando relevante, foco visível, contraste AA mínimo.

### Naming

| Contexto | Convenção |
|---|---|
| Display brand | `finance.sh` |
| Go module | `github.com/finance-sh/finance-sh` |
| Containers/networks Docker | `finance-sh-*` |
| Postgres ident (user/db) | `finance_sh` (snake) |
| Docker image (ghcr) | `finance-sh-app` |
| Email placeholder | `*@finance.sh` |
| localStorage keys | `finance-sh-*` |

---

## Testes

### Backend

```bash
cd backend
go test ./... -count=1 -race -short      # rápido (sem integration)
go test ./... -count=1 -race             # full (sobe testcontainers de Postgres)
```

Cobertura alvo: novos services com lógica não-trivial **devem** ter teste de unidade. Repositories podem ser cobertos por integration. Handlers cobertos por end-to-end via `httptest`.

### Frontend

```bash
cd frontend
npm run test            # vitest
npm run test:ui         # interativo
npm run lint            # ESLint + tsc --noEmit
npm run build           # smoke build (TS strict)
```

### Landing

```bash
cd landing
npm run build
```

PRs com testes quebrando **não são aceitos**. Use `git rebase` e corrija antes de pedir review.

---

## Mensagens de commit

Usamos **[Conventional Commits](https://www.conventionalcommits.org/)**.

```
<tipo>(<escopo>): <descrição em pt-BR ou en>

[corpo opcional explicando POR QUÊ, não o quê — o diff já mostra o quê]

[footer opcional: Closes #123, BREAKING CHANGE: ...]
```

Tipos aceitos:

- `feat` — nova funcionalidade
- `fix` — bugfix
- `refactor` — refactor sem mudar comportamento externo
- `perf` — performance
- `docs` — documentação
- `test` — testes adicionados/ajustados
- `chore` — build, deps, infra interna
- `ci` — CI/CD
- `i18n` — tradução

Exemplos:

```
feat(reports): exporta DRE em PDF com gráfico de fluxo
fix(auth): JWT refresh não invalidava token revogado
docs(contributing): adiciona seção sobre i18n
```

Commits squashed no merge — você não precisa rebasear seu PR para 1 commit, mantenedor faz no merge.

---

## Branches e PR

1. **Fork** o repo.
2. Branch a partir de `main`: `feat/<descricao-curta>` ou `fix/<descricao-curta>`. Evite `dev`, `wip`, `temp`.
3. Faça commits seguindo [convenção](#mensagens-de-commit).
4. Rode testes + lint antes de abrir PR.
5. Abra PR contra `main` com:
   - Título seguindo Conventional Commits (sem o footer).
   - Descrição com **contexto**, **motivação**, **como testei**, **screenshots** se for mudança visual.
   - Link da issue: `Closes #123` ou `Refs #123`.
6. CI precisa passar (build, test, lint). Se quebrar, push correção no mesmo branch.
7. Review: pelo menos 1 mantenedor aprova. Pode pedir mudanças — não leve pro pessoal, é sobre código.
8. Merge: estilo **squash and merge**. Histórico fica limpo.

### O que mantenedor procura

- Coerência com arquitetura existente.
- Testes cobrindo o caminho feliz + pelo menos 1 caminho de erro.
- Sem regressão de performance perceptível (use `go test -bench` se for hot path).
- i18n para qualquer string visível ao usuário (pt-BR, en, es no mínimo).
- Comentário em português ou inglês, consistente com o arquivo.
- LGPD: PII nova vai criptografada (`crypto.EncryptedString`)? Audit log captura? Justificar nos casos negativos.

---

## Traduções (i18n)

Idiomas suportados: **pt-BR** (primário), **en**, **es**.

Arquivos em `frontend/src/i18n/locales/<lang>/translation.json`.

Para adicionar idioma:

1. Crie pasta `frontend/src/i18n/locales/<codigo-iso>/`.
2. Copie `pt-BR/translation.json` como base e traduza.
3. Registre em `frontend/src/i18n/index.ts`.
4. Adicione opção em `frontend/src/components/LanguageSelect.tsx`.
5. PR rotulado `i18n`.

Mantenha a estrutura de chaves idêntica entre locales. Use `{{variavel}}` para interpolação.

---

## Licença e DCO

Este projeto é **AGPL-3.0**. Veja [LICENSE](LICENSE).

Implicações:

- Você pode usar, modificar e redistribuir.
- Se você **hospeda publicamente** uma versão modificada (mesmo via rede), precisa **disponibilizar o código-fonte modificado**.
- Forks comerciais privados são permitidos para uso interno.
- Sua contribuição é licenciada sob AGPL-3.0 automaticamente ao abrir PR.

### DCO (Developer Certificate of Origin)

Todo commit precisa ser **signed-off**:

```bash
git commit -s -m "feat(reports): exporta DRE em PDF"
```

O `-s` adiciona `Signed-off-by: Seu Nome <seu@email>` no footer, atestando que você tem direito de contribuir o código sob a licença. Bot do CI verifica.

Sem CLA (Contributor License Agreement). DCO é suficiente.

---

## Comunidade

- **Discussões e dúvidas:** [GitHub Discussions](https://github.com/finance-sh/finance-sh/discussions)
- **Bugs e features:** [Issues](https://github.com/finance-sh/finance-sh/issues)
- **Código de conduta:** [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) (em breve — base [Contributor Covenant 2.1](https://www.contributor-covenant.org/version/2/1/code_of_conduct/))

Seja respeitoso. Críticas são sobre código, não sobre pessoas. Mantenedores reservam o direito de remover comentários abusivos e banir contas reincidentes.

---

## Reconhecimento

Cada commit no `main` é creditado no histórico do Git. Contribuidores recorrentes são listados em `CONTRIBUTORS.md` (gerado por script). Mudanças destacadas vão para o `CHANGELOG.md`.

Obrigado por ajudar a deixar **finance.sh** melhor para quem prefere rodar localmente. 🇧🇷
