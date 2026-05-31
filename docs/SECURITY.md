# Segurança — finance.sh

Este documento descreve o modelo de ameaças, os controles implementados, a gestão
de segredos e os itens ainda pendentes (deferred). Linguagem: pt-BR.

> Reporte vulnerabilidades para **security@finance.sh** (placeholder). Não abra
> issues públicas para falhas de segurança.

---

## 1. Modelo de ameaças (resumo)

finance.sh é um controle financeiro **open-source self-hosted** (AGPL-3.0):
armazena dados financeiros e PII de múltiplas **organizações lógicas** dentro do
mesmo deploy do operador. Principais ativos e ameaças consideradas:

| Ativo | Ameaça | Controle principal |
|---|---|---|
| Credenciais de login | Força bruta, credential stuffing | bcrypt + lockout + rate limit |
| Sessões (JWT) | Roubo/replay de token | TTL curto + refresh rotacionado + revogação server-side |
| Dados de uma organização | Acesso cruzado entre organizações no mesmo deploy | Escopo obrigatório por `organization_id` + RBAC |
| PII / segredo 2FA | Vazamento do banco | Criptografia de campo (AES-256-GCM) |
| Perda da `ENCRYPTION_KEY` | Campos cifrados irrecuperáveis | Custódia da chave separada do banco + incluí-la no plano de backup |
| Tráfego cliente↔app | Interceptação (MITM) | TLS terminado no reverse proxy do operador + HSTS (quando atrás de TLS) |
| Postgres | Exposição na rede | Rede `internal`, sem rota externa, bind 127.0.0.1 |
| Container comprometido | Escalada de privilégio | no-new-privileges, cap_drop, read_only, non-root |
| Segredos no repo | Commit acidental | .gitignore + gitleaks no CI |
| Dependências | CVEs conhecidas | govulncheck, npm audit, trivy |

Fora de escopo neste estágio: DDoS volumétrico (deve ser mitigado por CDN/WAF
na borda do operador), ataques físicos ao host, ameaça interna com acesso root
ao host do operador.

> **Nota sobre multi-organização.** "Multi-org" significa **múltiplas
> organizações lógicas** dentro do mesmo deploy (família, sócios, contadores) —
> **não** clientes diferentes compartilhando infraestrutura. O isolamento é
> lógico (por `organization_id`).

---

## 2. Controles implementados

### Transporte (TLS)
- **A app fala HTTP puro** numa porta (default `8090`) e **não embute TLS**. O
  operador põe seu próprio **reverse proxy** (Traefik, Caddy, Nginx Proxy
  Manager, nginx) na frente pra terminar TLS — padrão do nicho self-hosted.
- O proxy deve repassar `X-Forwarded-Proto: https`; com isso a app emite
  **HSTS** `max-age=63072000; includeSubDomains` (middleware `SecurityHeaders`).
  Sem o header (acesso HTTP direto), HSTS não é enviado.
- Certificado real (Let's Encrypt) e redirect 80→443 ficam a cargo do proxy.
  Em dev, acesse `http://127.0.0.1:8090` direto. Ver README "Reverse proxy".

### Cabeçalhos HTTP / CSP (middleware `SecurityHeaders`)
Aplicados pelo backend (Go) em toda resposta (o binário serve a SPA diretamente):
- `X-Frame-Options: DENY`, `X-Content-Type-Options: nosniff`,
  `Referrer-Policy: strict-origin-when-cross-origin`,
  `Permissions-Policy: camera=(), microphone=(), geolocation=()`.
- **Content-Security-Policy** ajustada para SPA Vite (pulada só em `/swagger`,
  que carrega assets de CDN):
  `default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; font-src 'self' https://fonts.gstatic.com; img-src 'self' data:; connect-src 'self'; frame-ancestors 'none'; base-uri 'self'; form-action 'self'`.

### Autenticação e sessão
- Senhas com **bcrypt**.
- **JWT** par access/refresh; refresh **rotacionado** a cada uso e guardado como
  **hash** (revogável server-side). Access TTL curto (`JWT_ACCESS_TTL_MIN`).
- **Lockout de força bruta**: `LOGIN_MAX_ATTEMPTS` falhas bloqueiam por
  `LOGIN_LOCKOUT_MIN` minutos (estado persistido em Postgres).
- **Rate limiting** por minuto (`RATE_LIMIT_RPM`) — **in-memory por processo**
  do backend (token bucket via `golang.org/x/time/rate`), sem dependência
  externa. Em uma instância única basta; para múltiplas réplicas atrás de LB,
  o limite passa a ser por réplica.
- **2FA (TOTP)**: o segredo é armazenado **criptografado** (ver abaixo).

### Autorização e isolamento de organização
- **RBAC**: `owner` / `admin` / `member` / `viewer`.
- Toda entidade carrega `organization_id`; **todas as queries são escopadas** por
  ele. A organização ativa vem do header `X-Organization-ID`, validada contra a
  `Membership` do usuário.

### Criptografia de dados
- **Em campo**: PII sensível (ex.: notas com PII, segredo 2FA) cifrada com
  **AES-256-GCM** usando `ENCRYPTION_KEY` (base64 de 32 bytes).
- **⚠️ Custódia da `ENCRYPTION_KEY` (crítico).** A chave **não** está dentro do
  dump do banco — é o que decifra os campos cifrados. **Perder a chave =** esses
  campos ficam **permanentemente ilegíveis**, mesmo com o `pg_dump` intacto.
  Portanto:
  - Guarde a `ENCRYPTION_KEY` **separada do banco**, num gerenciador de segredos
    (Vault/KMS/Bitwarden) ou cofre offline. Tratar como chave-mestra.
  - **Nunca** trocar a chave depois de haver dados cifrados — não há rotação
    automática nesta versão; trocar invalida o que já foi cifrado.
  - Backup completo = **dump do banco + a `ENCRYPTION_KEY` + a `BACKUP_PASSPHRASE`**,
    guardados em locais distintos. Perder qualquer uma das chaves → perde o
    respectivo dado.
  - Em produção o app **recusa subir** com a chave de exemplo/dev (guard); gere
    uma única: `openssl rand -base64 32`.
- **Em repouso (disco)**: cabe **ao operador** manter o volume `pgdata` e o
  diretório de backups em **disco criptografado** (LUKS no homelab / volume
  criptografado na VPS). Isso cobre tudo que não é cifrado em campo (WAL,
  índices, dumps).
- **Backups**: `scripts/backup.sh` produz dumps **cifrados com GPG AES-256**
  (`--symmetric`), com poda por `RETENTION_DAYS`. Restauração via
  `scripts/restore.sh`. **Anexos de comprovante estão em BYTEA dentro do
  Postgres** (TOAST cuida da compressão out-of-line), portanto o `pg_dump`
  cobre os anexos automaticamente — não há object storage externo a
  preservar à parte.

### Auditoria e ciclo de vida do dado
- **Log de auditoria** para ações sensíveis.
- **Soft delete** (`deleted_at`) em todas as entidades.
- **PII mascarada nos logs** da aplicação.

### Isolamento de rede (Docker)
- Duas redes: **`finance-sh-edge`** (app — onde seu reverse proxy se conecta) e
  **`finance-sh-data`** com `internal: true` (postgres, app).
- Postgres **não tem rota externa** (só na rede interna); porta publicada ligada
  a **`127.0.0.1`**.
- A app publica HTTP na porta `APP_PORT`. Se o reverse proxy roda no mesmo host,
  fixe a publicação em `127.0.0.1` pra não expor a app direto na LAN.
- Stack default = **2 containers** (postgres + app); o que não existe não pode
  ser exposto. A SPA é embutida no binário (`go:embed`) — sem container
  frontend/nginx. Scheduler roda como goroutine in-process — sem container worker.

### Endurecimento de containers
- Aplicado nos **2 containers** default: postgres e app.
- `security_opt: no-new-privileges:true` em todos os serviços.
- `cap_drop: ALL` (a app não precisa de capabilities — binda porta alta como non-root).
- `read_only: true` + `tmpfs` nos caminhos graváveis (app: `/tmp`).
- Imagens da aplicação rodam **non-root** (uid 10001).
- `mem_limit` / `pids_limit` por serviço; **rotação de logs** json-file (10m × 3).
- Imagens fixadas por tag (TODO prod: fixar por **digest** `@sha256:`).

### CI / cadeia de suprimentos
- **govulncheck** (Go), **gosec** (SAST, advisory), **npm audit --audit-level=high**
  (frontend), **gitleaks** (segredos, histórico completo), **trivy** (filesystem +
  config + imagem). Build de imagem roda em todo PR (sem push); push só na `main`.

---

## 3. Gestão de segredos

Hoje (dev) os segredos vivem no `.env` (em texto, fora do Git via `.gitignore`).
**Isso não é aceitável em um deploy sério do operador (homelab/VPS/produção).**
Migração recomendada (cabe ao operador escolher):

**Opção A — Docker secrets** (Swarm/Compose):
```yaml
services:
  backend:
    secrets: [encryption_key, db_password]
    environment:
      ENCRYPTION_KEY_FILE: /run/secrets/encryption_key
secrets:
  encryption_key:
    external: true   # docker secret create encryption_key key.b64
  db_password:
    external: true
```
A app passa a ler `*_FILE` em vez do valor inline.

**Opção B — HashiCorp Vault / cloud KMS**:
- Segredos dinâmicos de banco (Vault database engine) e `ENCRYPTION_KEY` gerida
  por **KMS** (AWS KMS / GCP KMS / Azure Key Vault), com rotação e auditoria.
- Injeção via Vault Agent sidecar ou CSI Secrets Store no Kubernetes.

Regras gerais: nunca commitar segredos; rotacionar `JWT_*`, `ENCRYPTION_KEY` e
`DB_PASSWORD` antes de expor a instância para uso real; restringir quem lê o
`.env` no host do operador.

---

## 4. Itens pendentes (deferred)

Itens que o projeto **não** entrega de fábrica e que o operador deve avaliar
para um deploy sério, ou que os mantenedores planejam para versões futuras:

- [ ] **TLS / certificado real**: a app não embute TLS. O operador deve pôr um
      reverse proxy (Traefik/Caddy/Nginx Proxy Manager) na frente terminando
      HTTPS com Let's Encrypt. Exemplos no README "Reverse proxy". (Operador)
- [ ] **Postgres SSL no servidor** + `DB_SSLMODE=require`/`verify-full` (hoje
      `prefer` para garantir boot sem cert — ver §5). (Operador / projeto)
- [ ] **KMS** gerenciando `ENCRYPTION_KEY` (com rotação) em vez de `.env`.
      (Operador, se aplicável)
- [ ] **WAF** na borda (ModSecurity/Cloud WAF) e proteção DDoS. (Operador)
- [ ] **SIEM / alertas** (encaminhar audit log + logs de container para
      Loki/ELK + alertas). (Operador)
- [ ] **2FA com chaves de hardware** (WebAuthn/FIDO2) além de TOTP. (Projeto)
- [ ] Fixar **imagens por digest** e assinatura (cosign). (Projeto)
- [ ] Pentest externo antes de expor a instância publicamente. (Operador)

---

## 5. Decisão: `DB_SSLMODE=prefer` (não `require`) em dev

Habilitar SSL no `postgres:16-alpine` de forma confiável em um único passo exige
gerar/montar um cert de servidor com **permissão 600 e dono `postgres`** antes do
boot — frágil para um `docker compose up` limpo. Como o requisito é **"boot limpo
acima de estrito"**, o dev usa `DB_SSLMODE=prefer`: o backend negocia TLS se
disponível e cai para texto se não, **sempre subindo**. O caminho para `require`
+ cert de servidor está comentado no serviço `postgres` do `docker-compose.yml` e
listado como pendência acima.
