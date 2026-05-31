# Changelog

Todas as mudanças relevantes deste projeto são registradas aqui.

Formato baseado em [Keep a Changelog 1.1.0](https://keepachangelog.com/pt-BR/1.1.0/) e versionamento conforme [SemVer 2.0.0](https://semver.org/spec/v2.0.0.html).

> Convenções de tipo: **Added**, **Changed**, **Deprecated**, **Removed**, **Fixed**, **Security**.

---

## [0.1.0] — 2026-05-29

🎉 **Primeiro release** do finance.sh (AGPL-3.0). Controle financeiro 100% open-source self-hosted para pessoa física, MEI e microempresa no Brasil. Sem edições pagas, sem tiers.

### Added — Arquitetura

- **App em um único binário Go**: serve a SPA (embutida via `go:embed`) **e** a API JSON na mesma porta HTTP. Sem CORS em produção (mesma origem), sem container web dedicado.
- **Stack de 2 containers**: `app` + Postgres 16. **Única dependência externa é o Postgres** — anexos em BYTEA, cache/rate-limit in-memory no processo, lockout/refresh tokens no DB.
- **TLS por reverse proxy do operador** (Traefik/Caddy/Nginx Proxy Manager): a app publica HTTP puro em `APP_PORT` (default `8090`); o proxy termina TLS. Seção "Reverse proxy" no README com exemplos.
- **Scheduler in-process** (`JOBS_IN_PROCESS=true`): recorrência, notificações e purga LGPD rodam como goroutine dentro da app. `false` desliga (ex.: múltiplas réplicas onde só uma agenda).
- **Imagem multi-stage** (`backend/Dockerfile`, context = raiz): builda a SPA → embute no binário → runtime alpine non-root.

### Added — Autenticação e segurança

- **JWT** com access + refresh tokens e rotação. Refresh armazenado como hash, revogável.
- **bcrypt** para senha (cost 12). Validador de força mínima.
- **Lockout** de força bruta (`LOGIN_MAX_ATTEMPTS=5`, `LOGIN_LOCKOUT_MIN=15`).
- **2FA TOTP** (`pquerna/otp`) com **códigos de recuperação** únicos.
- **Verificação de e-mail** com token, fluxo de **esqueci minha senha** e **troca obrigatória de senha no primeiro acesso**.
- **Sessões** listáveis e revogáveis em Settings.
- **AES-256-GCM** para criptografia de campo (PII: notas de transação, segredo TOTP). Chave via `ENCRYPTION_KEY` (`openssl rand -base64 32`).
- **RBAC** granular: `owner`, `admin`, `member`, `viewer`.
- **Super-admin** de plataforma (back-office `/admin`) separado do RBAC por organização.
- **Admin de primeiro boot** (`BOOTSTRAP_ADMIN`, default true): banco vazio → cria super-admin + organização automaticamente, forçando troca de senha no 1º login. `ADMIN_PASSWORD` vazio gera senha aleatória forte e a imprime no log do boot.
- **Middleware de cabeçalhos de segurança** (Go): X-Frame-Options DENY, nosniff, Referrer-Policy, Permissions-Policy e **CSP** compatível com a SPA Vite em toda resposta; HSTS quando atrás de TLS (`X-Forwarded-Proto: https`).
- **Rate limiting** por IP (`RATE_LIMIT_RPM=120`, token bucket in-memory).
- **CORS** restrito por allowlist (`CORS_ORIGINS`).
- **Audit log** completo com UI dedicada. Captura método, path, ator, target.
- **Soft delete** em todas entidades + purga periódica via `RETENTION_DAYS`.
- **Hardening de container**: non-root, `no-new-privileges`, `cap_drop: ALL`, rootfs `read_only` + `tmpfs`, limites de memória e PIDs.
- **Isolamento de rede** Docker: Postgres em rede `internal` sem rota externa; porta publicada em `127.0.0.1`.

### Added — Multi-organização

- Mesma instância suporta **múltiplas organizações** (pessoal + PJ + família + sócios + contadores).
- Isolamento de dados por `organization_id` em todos repositórios.
- **Convites de membros** por e-mail com escolha de papel.
- Cabeçalho **`X-Organization-ID`** definindo tenant ativo da requisição.
- Múltiplas memberships por usuário — troca de org sem re-login.

### Added — Gestão financeira

- **Contas**: banco, carteira, investimento, cartão de crédito. Tipo, moeda, saldo inicial e ícone.
- **Transações**: receita, despesa, transferência entre contas. Vinculáveis a contatos, categorias, tags.
- **Categorias** com ícone e cor.
- **Contatos**: pagador/recebedor reutilizável (cliente, fornecedor, prestador).
- **Cartões de crédito** com **faturas** (ciclo fechamento → vencimento), pagamento de fatura como transação contabilizada, e **parcelamento** automático em N parcelas.
- **Contas a pagar/receber** com vencimento, baixa (settle) parcial ou total, e calendário consolidado.
- **Orçamentos** mensais por categoria, com indicador de consumo.
- **Metas** de economia com progresso e data alvo.
- **Recorrência** com regras (diária/semanal/mensal/anual + dia útil), materializada pelo scheduler in-process.
- **Tags** livres + **busca global** acessível por **⌘K / Ctrl-K**.
- **Categorização automática** por regras (descritivo → categoria).
- **Multi-moeda** por organização. 15 moedas ISO-4217 (BRL default).
- **Valores monetários sempre em centavos** (`int64`) — sem `float`.

### Added — Análise e dados

- **Dashboard** com cartões (saldo, receitas/despesas do mês, resultado), gráfico de **fluxo de caixa** (6 meses) e **top categorias** (donut).
- **Projeção de fluxo de caixa** para N meses à frente, com alerta de saldo negativo.
- **Relatórios** consolidados por período, com filtros por categoria, conta e tag.
- **Export** **Excel** (`excelize`), **PDF** (`go-pdf/fpdf`) e **CSV** com download direto.
- **Anexos** de comprovante (recibo, nota fiscal) em transações, armazenados em **Postgres BYTEA**. Limite por anexo `ATTACHMENT_MAX_MB=10`.
- **Import OFX/CSV** de extrato bancário com dedup + sugestão de categoria.

### Added — Plataforma

- **PWA** instalável (offline shell, auto-update via `vite-plugin-pwa`).
- **i18n** em 3 idiomas: **pt-BR** (primário), **en**, **es** via `react-i18next`.
- **Modo privacidade** que oculta valores no dashboard — toggle persistente.
- **Dark mode** com detecção de preferência do SO.
- **Setup wizard** de primeiro acesso (alternativa ao admin de boot, via `BOOTSTRAP_ADMIN=false`): detecta DB vazio, guia em 3 passos e cria o primeiro super-admin + organização em transação atômica. `GET /api/v1/setup/status` + `POST /api/v1/setup/initialize`.
- **Consentimento LGPD** versionado (`TERMS_VERSION`).
- **Exportar meus dados** (LGPD art. 18 II) em JSON e **excluir minha conta** com confirmação dupla + audit log.
- Banner de cookies com opt-in/opt-out persistente.

### Added — Open-source onboarding

- **Licença AGPL-3.0** (texto canônico GNU).
- **CONTRIBUTING.md** em pt-BR: setup dev, padrões Go + TS, Conventional Commits, DCO sign-off, workflow de PR.
- **CODE_OF_CONDUCT.md** bilíngue (Contributor Covenant 2.1, EN + pt-BR).
- **Issue templates** YAML (bug report, feature request, config) + **PR template** com checklist.
- **README.md** com screenshots e seção Reverse proxy.

### Added — DevOps

- **Subida com 1 comando**: `git clone && cp .env.example .env && docker compose up -d`. Defaults dev funcionam de cara.
- **Migrations versionadas** via `golang-migrate`.
- **Healthchecks** com dependências modeladas (`depends_on: service_healthy`).
- **Backup/restore** cifrados (`scripts/backup.sh` / `restore.sh`): `pg_dump` com GPG AES-256, retenção configurável. Como anexos vivem em BYTEA, o dump cobre tudo.
- **Makefile** com atalhos (`up`, `down`, `build`, `seed`, `backend-dev`, `frontend-dev`, `backup`, `restore`).
- **Logs JSON estruturados** (`slog`) com `request_id` propagado.
- **Swagger UI** em dev (`SWAGGER_ENABLED=true`), OpenAPI escrita à mão.

### Added — CI/CD

GitHub Actions workflow `.github/workflows/ci.yml` com:

- **Concurrency** cancel-in-progress por ref.
- **Build & test**: backend (gofmt, vet, build, test -race -short), frontend (lint + build TS strict).
- **Vuln scanning**: govulncheck (Go), gosec (SAST com SARIF upload), npm audit (fail em high+).
- **Secret scanning**: gitleaks (histórico inteiro).
- **Trivy** filesystem + misconfig com SARIF upload.
- **CodeQL** matriz `[go, javascript-typescript]`.
- **Imagem Docker** única `ghcr.io/<owner>/finance-sh-app` (tags `:latest` e `:sha-<7chars>` em pushes pra `main`) + Trivy image scan.

### Security

- AES-256-GCM para PII em campo (notas, segredo 2FA). Chave de produção via `openssl rand -base64 32`.
- bcrypt cost 12.
- JWT secrets separados pra access vs refresh; refresh rotaciona a cada uso e fica como hash no DB.
- Postgres com `DB_SSLMODE=prefer` em dev, `require`/`verify-full` em prod.
- Container `cap_drop: ALL` + `no-new-privileges`; imagem com user non-root (`finance_sh` UID 10001).
- Guarda de produção: a app se recusa a subir com secrets default (JWT/`ENCRYPTION_KEY`) quando `APP_ENV=production`.
- LGPD: `docs/LGPD.md` esclarece que em self-hosted o **operador** é o **controlador** dos dados. Ferramentas pra direitos do titular já incluídas.

### Notes

- **100% open-source, sem tiers.** Não há edição paga, plano, cobrança ou quota — todas as features estão neste repositório sob AGPL-3.0.
- **Sem provisionamento de cliente.** Cada usuário cria sua própria organização no cadastro (self-signup); membros entram por convite. O back-office super-admin é só leitura + operações de segurança (desativar usuário, resetar senha).
- A **landing page** (marketing) é um projeto/deploy à parte; não faz parte deste repositório.

---

[0.1.0]: https://github.com/finance-sh/finance-sh/releases/tag/v0.1.0
