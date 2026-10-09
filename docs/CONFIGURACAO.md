# Configuração — finance.sh

Todas as variáveis ficam no `.env` na raiz do projeto. Crie esse arquivo com
`./scripts/gen-env.sh`, que já gera os segredos; o modelo comentado é o
[`.env.example`](../.env.example). A coluna **Padrão** mostra o valor de uma
instalação via `docker compose`.

## App

| Variável | Padrão | Descrição |
|---|---|---|
| `APP_ENV` | `production` | `production` ou `development`. Em `production`, o app recusa `SEED=true`. |
| `APP_PORT` | `8090` | Porta HTTP publicada no host (dentro do container é `8080`). |
| `FRONTEND_URL` | `http://localhost:8090` | URL pública da app, usada nos links de e-mail. Em produção, a URL do seu proxy com HTTPS. |
| `SWAGGER_ENABLED` | `false` | Expõe a especificação e a UI de todos os endpoints. Ligue só em desenvolvimento. |
| `ATTACHMENT_MAX_MB` | `10` | Tamanho máximo de um comprovante anexado e de um extrato importado. |

## Primeiro acesso e contas

| Variável | Padrão | Descrição |
|---|---|---|
| `SETUP_TOKEN` | — | Código que o setup wizard pede (mínimo 16 caracteres; mais curto é ignorado). Vazio: um código aleatório a cada boot, impresso no log enquanto não há usuário. O definido aqui não é impresso. |
| `BOOTSTRAP_ADMIN` | `false` | `true` cria o admin no boot, sem o wizard (deploy headless). |
| `ADMIN_EMAIL` / `ADMIN_PASSWORD` / `ADMIN_ORG_NAME` | `admin@finance.sh` / vazio / `Minha Organização` | Usados só com `BOOTSTRAP_ADMIN=true`. Senha vazia: gerada e impressa no log. A troca é exigida no primeiro login. |
| `REGISTRATION_OPEN` | `true` | Cadastro público pela tela "Criar conta". Com `false`, só entram convidados e contas criadas pelo super-admin; um convidado sem conta depende do super-admin. Recomendado `false` numa instância exposta à internet. |
| `SEED` | `false` | Popula dados de demonstração no boot. Recusado com `APP_ENV=production`, porque as contas demo têm senhas públicas. |

## Banco de dados

| Variável | Padrão | Descrição |
|---|---|---|
| `DB_USER` / `DB_NAME` | `finance_sh` | Usuário e banco do Postgres. |
| `DB_PASSWORD` | gerada | Obrigatória; gerada pelo `gen-env.sh`. Trocar depois exige `ALTER ROLE` no banco. |
| `DB_SSLMODE` | `prefer` | Num banco remoto, use `require` ou `verify-full`. |
| `DB_PORT` / `PG_PORT` | `5433` | Porta do Postgres publicada só em `127.0.0.1`, para ferramentas locais. |
| `AUTO_MIGRATE` | — | `true` usa o AutoMigrate do GORM no lugar das migrações versionadas (só desenvolvimento). |

## Segredos e sessão

| Variável | Padrão | Descrição |
|---|---|---|
| `ENCRYPTION_KEY` | gerada | Chave AES-256 (base64, 32 bytes) que cifra as notas dos lançamentos e o segredo do 2FA. **Obrigatória**: o app não sobe sem ela, nem com a chave de dev publicada. Guarde fora do servidor (veja [Backups](INSTALACAO.md#faça-backup-da-encryption_key-crítico)). |
| `JWT_ACCESS_SECRET` / `JWT_REFRESH_SECRET` | gerados | Vazios: aleatórios a cada boot, o que desloga todo mundo num restart. |
| `JWT_ACCESS_TTL_MIN` / `JWT_REFRESH_TTL_DAYS` | `15` / `7` | Validade do access token e do refresh token. |
| `JWT_SESSION_MAX_DAYS` | `30` | Prazo máximo de uma sessão: renovar o refresh token não a estende além disso. Reapresentar um refresh token já trocado derruba a sessão inteira, porque indica token roubado. |

## Proteção contra abuso

| Variável | Padrão | Descrição |
|---|---|---|
| `LOGIN_MAX_ATTEMPTS` / `LOGIN_LOCKOUT_MIN` | `5` / `15` | Bloqueio por tentativas. No login e no 2FA, a contagem é por e-mail **e IP**, então quem só sabe o seu e-mail não tranca a sua conta. Na troca de senha e na exclusão de conta, é por conta. Fica em memória (um restart zera), e redefinir a senha zera tudo. |
| `RATE_LIMIT_RPM` | `120` | Requisições por minuto por IP na API. |
| `AUTH_RATE_LIMIT_RPM` | `30` | Limite mais apertado, por IP, para login, cadastro, 2FA, esqueci a senha, verificação de e-mail e `/setup/initialize`. Refresh e logout ficam no limite geral. |
| `TRUSTED_PROXIES` | só loopback | IPs ou CIDRs (separados por vírgula) cujo `X-Forwarded-For` / `X-Real-IP` é aceito; de qualquer outro endereço, o cabeçalho é ignorado. Proxy em outro container: ponha a rede dele (ex.: `172.16.0.0/12`), senão todos os clientes dividem o limite do IP do proxy. |
| `CORS_ORIGINS` | `http://localhost:8090,http://localhost:5173` | Origens permitidas, separadas por vírgula. Em produção a SPA é servida pela própria app, então só importa para o Vite em desenvolvimento. |

## E-mail (SMTP)

| Variável | Padrão | Descrição |
|---|---|---|
| `SMTP_HOST` / `SMTP_PORT` | vazio / `587` | Sem `SMTP_HOST`, os e-mails vão para o log, inclusive os links de redefinição de senha (veja o [aviso](INSTALACAO.md#esqueci-a-senha-sem-smtp)). Com SMTP, aceitar convite exige e-mail verificado. |
| `SMTP_USER` / `SMTP_PASS` | vazio | Credenciais do servidor de e-mail. |
| `SMTP_FROM` | `no-reply@finance.sh` | Remetente. |

## LGPD, tarefas e backup

| Variável | Padrão | Descrição |
|---|---|---|
| `RETENTION_DAYS` | `90` | Dias até a purga definitiva dos dados excluídos. |
| `TERMS_VERSION` | `1.0` | Versão dos Termos e da Política de Privacidade (o consentimento é versionado). |
| `JOBS_IN_PROCESS` | `true` | Roda o scheduler (recorrência, notificações, purga) dentro da app. `false` desliga, por exemplo quando há várias réplicas. |
| `WORKER_INTERVAL_SEC` | `3600` | Intervalo do scheduler, em segundos. |
| `BACKUP_PASSPHRASE` | — | Senha do GPG para `make backup` e `make restore`. Não fica no `.env` por padrão; exporte no shell. |
| `BACKUP_DIR` / `BACKUP_RETENTION_DAYS` | `./backups` / `90` | Onde os dumps cifrados ficam e por quantos dias. |

## Frontend (build)

| Variável | Padrão | Descrição |
|---|---|---|
| `VITE_API_URL` | `/api/v1` | Base da API usada pela SPA (mesma origem em produção). |
| `VITE_LANDING_URL` | vazio | URL de uma landing page externa, linkada na tela de login. Vazio esconde o link. |
