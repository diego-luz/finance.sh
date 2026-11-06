#!/usr/bin/env bash
# ============================================================================
# finance.sh — update an installation to the latest code, safely
#
# What it does, in order (and stops at the first problem):
#   1. Pre-flight: git checkout, Docker access, .env present, no local edits
#      to tracked files.
#   2. Fetches origin and shows what is about to change. Works even when the
#      upstream history was rewritten (where a plain `git pull` fails).
#   3. Warns about new .env keys and changes to docker-compose.yml or to the
#      database migrations, then asks for confirmation (skip with --yes).
#   4. Backs up the database BEFORE touching anything: encrypted through
#      scripts/backup.sh when BACKUP_PASSPHRASE is set, otherwise a plain
#      pg_dump (-Fc, mode 600) in BACKUP_DIR/pre-update.
#   5. Tags the running app image as `<image>:pre-update`, for rollback.
#   6. Moves the code to origin/<branch> and rebuilds/recreates the app.
#      Postgres and its volume are not touched.
#   7. Waits for the app healthcheck. If the build fails, the old container is
#      still running and the code goes back to the previous commit. If the new
#      container does not get healthy, it rolls back automatically to the
#      previous image and commit.
#   8. Compares row counts per table before/after, so you can see at a glance
#      that no data went missing.
#
# Usage (from the installation directory, as a user that can write to it and
# talk to Docker — usually: sudo):
#   sudo ./scripts/update.sh            # interactive
#   sudo ./scripts/update.sh --yes      # no questions (cron / automation)
#   sudo ./scripts/update.sh --check    # only shows whether there is an update
#   make update
#
# Options:
#   -y, --yes           do not ask for confirmation
#   -c, --check         fetch and report; change nothing
#   -b, --branch NAME   branch to follow (default: main)
#       --rebuild       rebuild even when the code is already up to date
#       --skip-backup   do not dump the database first (NOT recommended)
#   -h, --help          this help
#
# Optional (from .env or env):
#   DB_USER, DB_NAME, PG_CONTAINER, APP_CONTAINER, BACKUP_DIR,
#   BACKUP_PASSPHRASE, UPDATE_KEEP_BACKUPS (default 10 plain pre-update dumps),
#   UPDATE_HEALTH_TIMEOUT (default 180 seconds)
# ============================================================================
set -euo pipefail

# This file is part of the checkout it updates: `git reset` below rewrites it
# while bash is still reading it (bash reads scripts incrementally), which can
# run a mix of the old and new versions. So run from a private copy instead.
if [ -z "${FINANCE_SH_UPDATE_ROOT:-}" ]; then
  ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
  self_copy="$(mktemp "${TMPDIR:-/tmp}/finance-sh-update.XXXXXX")"
  cp "${BASH_SOURCE[0]}" "$self_copy"
  FINANCE_SH_UPDATE_ROOT="$ROOT_DIR" FINANCE_SH_UPDATE_SELF="$self_copy" \
    exec bash "$self_copy" "$@"
fi
ROOT_DIR="$FINANCE_SH_UPDATE_ROOT"
SCRIPT_DIR="$ROOT_DIR/scripts"
SELF_COPY="${FINANCE_SH_UPDATE_SELF:-}"
CMD="./scripts/update.sh"
TMP_FILES=("$SELF_COPY")
trap 'rm -f "${TMP_FILES[@]}"' EXIT
cd "$ROOT_DIR"

# shellcheck source=scripts/env.sh
. "$SCRIPT_DIR/env.sh"
load_env "$ROOT_DIR/.env" DB_USER DB_NAME PG_CONTAINER APP_CONTAINER BACKUP_DIR \
  BACKUP_PASSPHRASE UPDATE_KEEP_BACKUPS UPDATE_HEALTH_TIMEOUT

DB_USER="${DB_USER:-finance_sh}"
DB_NAME="${DB_NAME:-finance_sh}"
PG_CONTAINER="${PG_CONTAINER:-finance-sh-postgres}"
APP_CONTAINER="${APP_CONTAINER:-finance-sh-app}"
BACKUP_DIR="${BACKUP_DIR:-$ROOT_DIR/backups}"
KEEP_BACKUPS="${UPDATE_KEEP_BACKUPS:-10}"
HEALTH_TIMEOUT="${UPDATE_HEALTH_TIMEOUT:-180}"

ASSUME_YES=false
CHECK_ONLY=false
FORCE_REBUILD=false
SKIP_BACKUP=false
BRANCH=main

usage() { sed -n '/^# Usage/,/^# Optional/p' "$0" | sed '$d; s/^# \{0,1\}//'; }

while [ $# -gt 0 ]; do
  case "$1" in
    -y|--yes) ASSUME_YES=true ;;
    -c|--check) CHECK_ONLY=true ;;
    -b|--branch) BRANCH="${2:?--branch precisa de um nome}"; shift ;;
    --rebuild) FORCE_REBUILD=true ;;
    --skip-backup) SKIP_BACKUP=true ;;
    -h|--help) usage; exit 0 ;;
    *) echo "opção desconhecida: $1 (veja --help)" >&2; exit 2 ;;
  esac
  shift
done

# ---- Output helpers ----------------------------------------------------------
if [ -t 1 ]; then
  B=$'\e[1m'; G=$'\e[32m'; Y=$'\e[33m'; R=$'\e[31m'; D=$'\e[2m'; N=$'\e[0m'
else
  B=''; G=''; Y=''; R=''; D=''; N=''
fi
step() { printf '\n%s==> %s%s\n' "$B" "$*" "$N"; }
info() { printf '    %s\n' "$*"; }
ok()   { printf '    %s✔%s %s\n' "$G" "$N" "$*"; }
warn() { printf '    %s!%s %s\n' "$Y" "$N" "$*"; }
die()  { printf '\n%serro:%s %s\n' "$R" "$N" "$*" >&2; exit 1; }

# root owns the checkout on most servers; git refuses "dubious ownership"
# otherwise. Scoped to this command only, never written to any git config.
git() { command git -c safe.directory="$ROOT_DIR" "$@"; }

# ---- 1. Pre-flight -----------------------------------------------------------
step "Verificando a instalação em $ROOT_DIR"

git rev-parse --is-inside-work-tree >/dev/null 2>&1 \
  || die "$ROOT_DIR não é um repositório git (o update depende do git)."
command -v docker >/dev/null 2>&1 || die "docker não encontrado."
docker compose version >/dev/null 2>&1 || die "Docker Compose v2 não encontrado (docker compose)."
docker info >/dev/null 2>&1 \
  || die "sem acesso ao Docker com este usuário. Rode com sudo: sudo $CMD"
[ -w "$ROOT_DIR/.git" ] \
  || die "sem permissão de escrita em $ROOT_DIR. Rode com sudo: sudo $CMD"
[ -f "$ROOT_DIR/.env" ] || die ".env não encontrado. Numa instalação nova, rode ./scripts/gen-env.sh."

if [ -n "$(git status --porcelain --untracked-files=no)" ]; then
  git status --short --untracked-files=no | sed 's/^/      /'
  die "há alterações locais em arquivos do repositório (acima).
      O update sobrescreveria essas mudanças. Guarde-as (git stash) ou desfaça-as
      (git checkout -- <arquivo>) e rode de novo."
fi
ok "repositório limpo, Docker acessível, .env presente"

# ---- 2. What is about to change ----------------------------------------------
step "Buscando atualizações (origin/$BRANCH)"
git fetch --quiet origin "$BRANCH" || die "não consegui buscar origin/$BRANCH."

OLD="$(git rev-parse HEAD)"
NEW="$(git rev-parse "origin/$BRANCH")"
short() { git rev-parse --short "$1"; }

if [ "$OLD" = "$NEW" ]; then
  ok "já está na versão mais recente: $(git log -1 --format='%h %s' "$NEW")"
  if $CHECK_ONLY || ! $FORCE_REBUILD; then
    $FORCE_REBUILD || info "${D}Para reconstruir mesmo assim: sudo $CMD --rebuild${N}"
    exit 0
  fi
else
  info "atual: $(git log -1 --format='%h %s (%cs)' "$OLD")"
  info "nova:  $(git log -1 --format='%h %s (%cs)' "$NEW")"
  if git merge-base --is-ancestor "$OLD" "$NEW"; then
    COUNT="$(git rev-list --count --no-merges "$OLD..$NEW")"
    printf '\n'
    info "$COUNT commit(s) novos:"
    git log --format='%h %s' --no-merges "$OLD..$NEW" | head -25 | sed 's/^/      /'
    [ "$COUNT" -gt 25 ] && info "      ..."
  else
    warn "o histórico do repositório foi reescrito desde a sua versão;"
    warn "um 'git pull' falharia, mas este script lida com isso."
  fi
  printf '\n'
  info "arquivos alterados:$(git diff --shortstat "$OLD" "$NEW")"
fi

# ---- 3. Things that need a human look ----------------------------------------
NEEDS_ATTENTION=false
if [ "$OLD" != "$NEW" ]; then
  if ! git diff --quiet "$OLD" "$NEW" -- docker-compose.yml; then
    warn "o docker-compose.yml mudou. Se você usa um docker-compose.override.yml,"
    warn "confira se ele continua compatível:  git diff $(short "$OLD") $(short "$NEW") -- docker-compose.yml"
    NEEDS_ATTENTION=true
  fi
  if ! git diff --quiet "$OLD" "$NEW" -- backend/migrations; then
    warn "há migrações novas no banco; elas rodam sozinhas quando o app sobe."
    NEEDS_ATTENTION=true
  fi
  # Keys present in the NEW .env.example but absent from .env. Missing keys
  # fall back to the compose defaults, so this is advice, not an error.
  missing="$(comm -23 \
    <(git show "$NEW:.env.example" | sed -nE 's/^([A-Z][A-Z0-9_]*)=.*/\1/p' | sort -u) \
    <(sed -nE 's/^[[:space:]]*([A-Z][A-Z0-9_]*)=.*/\1/p' .env | sort -u))"
  if [ -n "$missing" ]; then
    warn "variáveis do .env.example que não estão no seu .env (usam o valor padrão):"
    printf '%s\n' "$missing" | sed 's/^/        /'
    NEEDS_ATTENTION=true
  fi
  $NEEDS_ATTENTION || ok "sem mudanças de configuração, compose ou banco"
fi

if $CHECK_ONLY; then
  printf '\n%sHá atualização disponível.%s Para aplicar: %s\n' "$B" "$N" "sudo $CMD"
  exit 0
fi

if ! $ASSUME_YES; then
  if [ ! -t 0 ]; then
    die "sem terminal para confirmar. Use --yes para rodar sem perguntas."
  fi
  printf '\n'
  read -r -p "Aplicar a atualização agora? [s/N] " answer
  case "$answer" in s|S|sim|SIM|y|Y|yes) ;; *) echo "Nada foi alterado."; exit 0 ;; esac
fi

# ---- Row counts (before/after) -----------------------------------------------
# One query, every table in public: the count is computed per table through
# query_to_xml, so no table list is hardcoded here.
row_counts() {
  docker exec "$PG_CONTAINER" psql -U "$DB_USER" -d "$DB_NAME" -AtF ' ' -c "
    SELECT table_name,
           (xpath('/row/c/text()', query_to_xml(
              format('SELECT count(*) AS c FROM %I.%I', table_schema, table_name),
              false, true, '')))[1]::text
      FROM information_schema.tables
     WHERE table_schema = 'public' AND table_type = 'BASE TABLE'
     ORDER BY 1"
}

docker ps --format '{{.Names}}' | grep -qx "$PG_CONTAINER" \
  || die "o container do banco ($PG_CONTAINER) não está rodando. Suba o stack antes: docker compose up -d"

BEFORE_FILE="$(mktemp)"; AFTER_FILE="$(mktemp)"
TMP_FILES+=("$BEFORE_FILE" "$AFTER_FILE")
row_counts | LC_ALL=C sort > "$BEFORE_FILE" || die "não consegui ler o banco ($PG_CONTAINER)."

# ---- 4. Backup ---------------------------------------------------------------
BACKUP_FILE=""
if $SKIP_BACKUP; then
  step "Backup"
  warn "pulado (--skip-backup)."
else
  step "Fazendo backup do banco antes de mexer em qualquer coisa"
  if [ -n "${BACKUP_PASSPHRASE:-}" ]; then
    "$SCRIPT_DIR/backup.sh" | sed 's/^/    /' \
      || die "o backup criptografado falhou; nada foi alterado."
    BACKUP_FILE="$(ls -1t "$BACKUP_DIR"/finance_sh-*.sql.gpg 2>/dev/null | head -1 || true)"
    [ -n "$BACKUP_FILE" ] || die "o backup não gerou arquivo; nada foi alterado."
    ok "backup criptografado: $BACKUP_FILE"
  else
    dir="$BACKUP_DIR/pre-update"
    mkdir -p "$dir"; chmod 700 "$dir"
    BACKUP_FILE="$dir/finance_sh-$(date +%Y%m%d-%H%M%S)-$(short "$OLD").dump"
    part="$BACKUP_FILE.part"
    ( umask 077
      docker exec "$PG_CONTAINER" pg_dump -U "$DB_USER" -d "$DB_NAME" -Fc > "$part" ) \
      || { rm -f "$part"; die "pg_dump falhou; nada foi alterado."; }
    # A dump that pg_restore can list is a dump that is complete.
    docker exec -i "$PG_CONTAINER" pg_restore -l < "$part" >/dev/null \
      || { rm -f "$part"; die "o backup gerado não é válido; nada foi alterado."; }
    mv "$part" "$BACKUP_FILE"
    ok "backup: $BACKUP_FILE ($(du -h "$BACKUP_FILE" | cut -f1))"
    warn "sem BACKUP_PASSPHRASE no .env, este backup NÃO é criptografado (só o"
    warn "dono do arquivo lê). Para cifrar, defina BACKUP_PASSPHRASE; veja docs/INSTALACAO.md."
    # Keep the newest N plain pre-update dumps.
    ls -1t "$dir"/finance_sh-*.dump 2>/dev/null | tail -n +"$((KEEP_BACKUPS + 1))" | xargs -r rm -f || true
  fi
fi

# ---- 5. Remember the running image, for rollback -----------------------------
IMAGE="$(docker compose config --images 2>/dev/null | grep -E -- '-app(:|$)' | head -1 || true)"
IMAGE="${IMAGE:-finance-sh-app}"
IMAGE="${IMAGE%%:*}"
ROLLBACK_IMAGE=""
if old_id="$(docker inspect --format '{{.Image}}' "$APP_CONTAINER" 2>/dev/null)"; then
  docker tag "$old_id" "$IMAGE:pre-update"
  ROLLBACK_IMAGE="$IMAGE:pre-update"
fi

# ---- 6. Update code and rebuild ----------------------------------------------
step "Atualizando o código para $(short "$NEW")"
git reset --quiet --hard "$NEW"
ok "código em $(git log -1 --format='%h %s')"

step "Reconstruindo e reiniciando o app (pode levar alguns minutos)"
info "${D}o Postgres e os dados não são recriados; só o container do app${N}"
if ! docker compose up -d --build; then
  git reset --quiet --hard "$OLD"
  die "o build falhou. O app antigo continua rodando e o código voltou para $(short "$OLD").
      Nenhum dado foi alterado. Veja a saída acima."
fi

# ---- 7. Health ---------------------------------------------------------------
step "Esperando o app ficar saudável (até ${HEALTH_TIMEOUT}s)"
health=""
for _ in $(seq 1 "$((HEALTH_TIMEOUT / 3))"); do
  health="$(docker inspect --format '{{if .State.Health}}{{.State.Health.Status}}{{else}}{{.State.Status}}{{end}}' "$APP_CONTAINER" 2>/dev/null || echo missing)"
  case "$health" in healthy) break ;; unhealthy|exited|dead|missing) break ;; esac
  sleep 3
done

if [ "$health" != healthy ]; then
  warn "o app não ficou saudável (estado: ${health:-desconhecido}). Últimas linhas do log:"
  docker logs --tail 25 "$APP_CONTAINER" 2>&1 | sed 's/^/      /'
  if [ -n "$ROLLBACK_IMAGE" ]; then
    step "Voltando para a versão anterior"
    git reset --quiet --hard "$OLD"
    docker tag "$ROLLBACK_IMAGE" "$IMAGE:latest"
    docker compose up -d --no-build --force-recreate app
    die "a atualização foi desfeita: código em $(short "$OLD") e a imagem anterior rodando.
      Backup feito antes da tentativa: ${BACKUP_FILE:-nenhum}"
  fi
  die "o app não subiu e não havia imagem anterior para voltar. Backup: ${BACKUP_FILE:-nenhum}"
fi
ok "app saudável"

# ---- 8. Data check -----------------------------------------------------------
step "Conferindo os dados"
row_counts | LC_ALL=C sort > "$AFTER_FILE" || die "não consegui ler o banco depois da atualização."
# Tables that churn on their own (sessions, one-time codes, retention purge)
# are reported but never counted as data loss.
VOLATILE='^(refresh_tokens|password_resets|email_verifications|notifications|audit_logs|invitations|schema_migrations)$'
lost=0; changed=0
while read -r table before after; do
  [ "$before" = "$after" ] && continue
  changed=$((changed + 1))
  if [ "$after" = "-" ]; then
    warn "$table: $before linhas antes, tabela não existe mais"
    lost=$((lost + 1))
  elif [ "$before" = "-" ]; then
    info "$table: tabela nova ($after linhas)"
  elif [ "$after" -lt "$before" ] && ! [[ $table =~ $VOLATILE ]]; then
    warn "$table: $before → $after linhas"
    lost=$((lost + 1))
  else
    info "${D}$table: $before → $after (normal)${N}"
  fi
done < <(LC_ALL=C join -a1 -a2 -e - -o 0,1.2,2.2 "$BEFORE_FILE" "$AFTER_FILE")

total="$(awk '{s+=$2} END {print s+0}' "$AFTER_FILE")"
tables="$(wc -l < "$AFTER_FILE")"
if [ "$lost" -eq 0 ]; then
  ok "$tables tabelas, $total registros; nenhum dado a menos"
else
  warn "$lost tabela(s) com menos registros que antes (acima). O backup está em:"
  warn "  ${BACKUP_FILE:-nenhum (--skip-backup)}"
fi

# Old images left dangling by the rebuild; the :pre-update tag stays.
docker image prune -f >/dev/null 2>&1 || true

# ---- Summary -----------------------------------------------------------------
printf '\n%s%sAtualização concluída.%s\n' "$B" "$G" "$N"
info "versão:  $(short "$OLD") → $(git log -1 --format='%h %s')"
info "backup:  ${BACKUP_FILE:-nenhum (--skip-backup)}"
if [ -n "$ROLLBACK_IMAGE" ]; then
  info "${D}para voltar à versão anterior:${N}"
  info "${D}  git -c safe.directory=$ROOT_DIR reset --hard $(short "$OLD") && docker tag $ROLLBACK_IMAGE $IMAGE:latest && docker compose up -d --no-build --force-recreate app${N}"
fi
[ "$lost" -eq 0 ] || exit 1
