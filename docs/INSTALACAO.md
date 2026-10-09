# Instalação e operação — finance.sh

Como subir, acessar pela primeira vez, expor com TLS, fazer backup e
recuperar acesso. Variáveis de ambiente estão em
[`CONFIGURACAO.md`](CONFIGURACAO.md).

## Sumário

- [Início rápido](#início-rápido)
- [Primeiro acesso](#primeiro-acesso)
- [Desenvolvimento local sem Docker](#desenvolvimento-local-sem-docker)
- [Reverse proxy (TLS)](#reverse-proxy-tls)
- [Backups criptografados](#backups-criptografados)
- [Esqueci a senha (sem SMTP)](#esqueci-a-senha-sem-smtp)

---

## Início rápido

**Pré-requisitos:** Docker 24+ e Docker Compose v2.

```bash
# 1. Clone
git clone https://github.com/diego-luz/finance.sh && cd finance.sh

# 2. Crie o .env com segredos gerados para ESTA instalação
./scripts/gen-env.sh          # (ou: make env)

# 3. Suba a stack inteira
docker compose up -d --build
```

> **Por que um script, e não `cp .env.example .env`:** o repositório não
> versiona nenhum segredo que funcione, porque um valor commitado é um valor
> que todo mundo que leu o projeto já tem. O script gera `ENCRYPTION_KEY`,
> `DB_PASSWORD` e os segredos de JWT na hora. O app se recusa a subir sem uma
> `ENCRYPTION_KEY` própria, **em qualquer ambiente**, e não só com
> `APP_ENV=production`, porque essa é justamente a variável que se esquece de
> trocar. Guarde a `ENCRYPTION_KEY` num cofre: sem ela, os campos
> criptografados ficam ilegíveis.

Depois de subir (~1 min no primeiro build), ficam ativos **2 containers**:

| Container | Função |
|---|---|
| `finance-sh-postgres` | PostgreSQL 16 |
| `finance-sh-app` | Binário Go: SPA embutida (`go:embed`) + API + scheduler in-process, numa porta HTTP |

| Serviço | URL | Observação |
|---|---|---|
| **App (SPA + API)** | **http://127.0.0.1:8090** | HTTP puro. Em produção, fica atrás do seu proxy com TLS. |
| Swagger UI | http://127.0.0.1:8090/swagger | Só com `SWAGGER_ENABLED=true`. |
| PostgreSQL | `127.0.0.1:5433` | Só loopback. Ou `docker exec finance-sh-postgres psql -U finance_sh`. |

Para popular dados de demonstração, ponha `SEED=true` e
`APP_ENV=development` no `.env`, ou rode `make seed` (o padrão é `SEED=false`).
Com `APP_ENV=production`, o app se recusa a subir com `SEED=true`, porque as
contas de demonstração têm senhas públicas.

---

## Primeiro acesso

Com o **banco vazio**, a app abre o **setup wizard**, onde **você cria** o
primeiro super-admin (nome, e-mail, senha) e a organização. Nenhuma senha é
exibida: você mesmo a define. Esse é o padrão (`BOOTSTRAP_ADMIN=false`).

1. Suba a stack e pegue o **código de instalação** no log: `docker compose logs app`.
2. Abra **http://127.0.0.1:8090**. Com o banco vazio, a app cai em **/setup**.
3. Informe o código, preencha admin e organização, e pronto: já entra logado.

```
┌────────────────────────────────────────────────────────────┐
│ PRIMEIRO ACESSO — código de instalação                     │
│                                                            │
│   K7QM-2XRA-PZ4W-ND6T  (novo a cada reinício)              │
└────────────────────────────────────────────────────────────┘
```

O código prova que quem está no navegador também tem acesso ao servidor. Sem
ele, quem chegasse primeiro a uma instância recém-exposta viraria o
super-admin. O código muda a cada reinício; para fixá-lo (deploy
automatizado), defina `SETUP_TOKEN`. No servidor, o setup também é protegido
por `users-count == 0` numa transação com lock: só roda uma vez, ninguém
recria o admin depois, e duas chamadas simultâneas não criam dois.

> **Não há "admin separado".** O usuário criado no wizard é ao mesmo tempo
> **super-admin da plataforma** (back-office `/admin`) e **dono (owner)** da
> primeira organização. As contas `super@finance.sh` / `admin@finance.sh` que
> aparecem por aí são só do **seed de dev** (`SEED=true`) e do **bootstrap
> headless** (`BOOTSTRAP_ADMIN=true`); não existem num deploy real.

### Deploy headless (sem UI), opcional

Para automação ou CI sem navegador, use `BOOTSTRAP_ADMIN=true`. A app cria o
admin no boot e, se `ADMIN_PASSWORD` estiver vazio, **gera uma senha e a
imprime no log**:

```
┌────────────────────────────────────────────────────────────┐
│ ADMIN CRIADO — troque a senha no 1º login                    │
│   email:  admin@finance.sh                                   │
│   senha:  7Kq9-mZ2x-Vp4w-Rt6n  (aleatória)                   │
└────────────────────────────────────────────────────────────┘
```

> A senha aparece em `docker compose logs app`. Use esse modo só onde você
> controla quem lê o log, e não exponha a porta antes do primeiro login.

### Credenciais de demonstração (só dev)

Com `SEED=true` (só com `APP_ENV=development`), o backend cria usuários de exemplo:

| Tipo | E-mail | Senha |
|---|---|---|
| Usuário comum (org demo) | `demo@finance.sh` | `senha123` |
| Super-admin (plataforma) | `super@finance.sh` | `superadmin123` |

O super-admin do seed é obrigado a trocar a senha no primeiro login.

---

## Desenvolvimento local sem Docker

Sobe só o banco via Compose e roda o app na máquina (loop rápido):

```bash
docker compose up -d postgres                # única dependência
make backend-dev                             # cd backend && go run ./cmd/api  (:8090)
make frontend-dev                            # cd frontend && npm run dev      (:5173)
```

O Vite (`5173`) faz proxy de `/api` para `http://localhost:8090`. O Postgres
fica na porta `5433` do host. Em dev você usa o Vite com HMR; o binário Go só
serve a SPA embutida nos builds Docker e de produção. Mais detalhes em
[`CONTRIBUTING.md`](../CONTRIBUTING.md).

---

## Reverse proxy (TLS)

A app publica **HTTP puro** em `APP_PORT` (padrão `8090`) e **não embute TLS**,
de propósito. Em self-hosted e homelab você quase sempre já tem um reverse
proxy cuidando do Let's Encrypt; basta apontá-lo para a app. É o mesmo modelo
do Vaultwarden, do Miniflux e do Paperless-ngx.

- **Proxy no mesmo host:** fixe a porta em loopback no `docker-compose.yml`
  (`"127.0.0.1:${APP_PORT:-8090}:8080"`), para a app não ficar exposta direto.
- **Proxy em outro container ou outro host:** ponha a rede dele em
  `TRUSTED_PROXIES` (ex.: `172.16.0.0/12` na rede do Docker). Sem isso, a app
  ignora o `X-Forwarded-For`, e todos os clientes dividem o limite de
  requisições do IP do proxy. O log avisa quando detecta esse caso.
- Defina `FRONTEND_URL` com a URL pública (ex.: `https://finance.example.com`),
  para os links de e-mail saírem certos.

**Caddy** (TLS automático, `Caddyfile`):

```caddy
finance.example.com {
    reverse_proxy 127.0.0.1:8090
}
```

**Traefik** (labels no compose; ligue a app à rede do Traefik):

```yaml
labels:
  - "traefik.enable=true"
  - "traefik.http.routers.finance.rule=Host(`finance.example.com`)"
  - "traefik.http.routers.finance.tls.certresolver=le"
  - "traefik.http.services.finance.loadbalancer.server.port=8080"
```

**Nginx Proxy Manager:** crie um Proxy Host apontando para o host da app,
porta `8090`, e peça o certificado Let's Encrypt na aba SSL.

**Nginx manual:**

```nginx
server {
    server_name finance.example.com;
    listen 443 ssl;        # certificados via certbot
    location / {
        proxy_pass http://127.0.0.1:8090;
        proxy_set_header Host $host;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

---

## Backups criptografados

```bash
export BACKUP_PASSPHRASE='uma-passphrase-forte'   # fora do histórico do shell
make backup                                       # gera backups/finance_sh-...sql.gpg (GPG AES-256)
make restore FILE=backups/finance_sh-AAAAMMDD-HHMMSS.sql.gpg
```

- **Backup:** os dumps são cifrados com **GPG AES-256** e apagados depois de
  `BACKUP_RETENTION_DAYS` (padrão 90). O arquivo só é gravado quando o dump
  termina sem erro.
- **Restore:** roda numa transação única, que para no primeiro erro, e
  desliga o container da app durante a carga.
- **Disco:** guarde o diretório de backups e o volume `pgdata` em **disco
  criptografado** (LUKS ou volume cloud criptografado). É isso que protege em
  repouso o que não é cifrado campo a campo.

### Faça backup da `ENCRYPTION_KEY` (crítico)

As notas dos lançamentos e o segredo do 2FA são cifrados com a
`ENCRYPTION_KEY`, e **ela não está no dump do banco**. Se você perder a chave,
esses dados ficam **permanentemente ilegíveis**, mesmo com o backup intacto.

- Guarde a chave **separada do banco**, num gerenciador de segredos
  (Bitwarden, Vault, 1Password) ou num cofre offline. Trate-a como uma senha
  mestra.
- **Nunca commite** a chave; o `.env` está no `.gitignore` de propósito.
- **Não troque a chave** depois de ter dados cifrados: não há rotação
  automática, e trocar significa perder o que já foi cifrado.

> Regra prática: **2 segredos para guardar fora do servidor**, a
> `ENCRYPTION_KEY` (decifra os campos) e a `BACKUP_PASSPHRASE` (abre os
> dumps). Perdeu qualquer um, perdeu o dado correspondente.

---

## Esqueci a senha (sem SMTP)

Self-hosted costuma rodar **sem SMTP**. Há 3 caminhos para recuperar acesso:

1. **CLI (operador com shell), o recomendado.** Redefine a senha de qualquer
   usuário direto do servidor:

   ```bash
   docker compose run --rm app -reset-password admin@finance.sh
   ```

   Gera uma senha aleatória, **imprime no terminal**, força a troca no próximo
   login e revoga as sessões ativas.

2. **Link no log.** O "esqueci a senha" da tela gera um token. **Sem SMTP, o
   link `…/reset-password?token=…` vai para o log**: pegue-o em
   `docker compose logs app` e abra no navegador.

   > **Atenção:** sem SMTP, quem lê o log da app consegue redefinir a senha de
   > **qualquer** conta, inclusive a do super-admin. Restrinja o acesso aos
   > logs como restringe o acesso ao servidor, ou configure o SMTP.

3. **Super-admin.** No back-office `/admin` → Usuários, redefina a senha de
   outro usuário. Ele é obrigado a trocá-la no próximo login.

Com SMTP configurado (`SMTP_*`), o "esqueci a senha" envia o link por e-mail.
