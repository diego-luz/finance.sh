#!/usr/bin/env bash
# ============================================================================
# finance.sh — restore an encrypted Postgres backup
#
# Decrypts a *.sql.gpg file produced by scripts/backup.sh and pipes it into
# psql in the running `finance-sh-postgres` container. THIS OVERWRITES the target
# database (the dump was taken with --clean --if-exists).
#
# COBERTURA: anexos de comprovante vivem em BYTEA dentro do Postgres, então
# restaurar o dump devolve os anexos junto. Não há object storage externo
# para restaurar à parte.
#
# Usage:
#   ./scripts/restore.sh backups/finance_sh-finance_sh-20260526-120000.sql.gpg
#   make restore FILE=backups/finance_sh-finance_sh-20260526-120000.sql.gpg
#
# Required (from .env or env):
#   DB_USER, DB_NAME
#   BACKUP_PASSPHRASE          same passphrase used to create the backup
# Optional:
#   PG_CONTAINER (default finance-sh-postgres)
# ============================================================================
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

# shellcheck source=scripts/env.sh
. "$SCRIPT_DIR/env.sh"
load_env "$ROOT_DIR/.env" DB_USER DB_NAME PG_CONTAINER APP_CONTAINER BACKUP_PASSPHRASE

FILE="${1:-${FILE:-}}"
if [ -z "$FILE" ] || [ ! -f "$FILE" ]; then
    echo "ERROR: pass the encrypted backup file to restore." >&2
    echo "Usage: $0 <path-to.sql.gpg>" >&2
    exit 1
fi

DB_USER="${DB_USER:-finance_sh}"
DB_NAME="${DB_NAME:-finance_sh}"
PG_CONTAINER="${PG_CONTAINER:-finance-sh-postgres}"
APP_CONTAINER="${APP_CONTAINER:-finance-sh-app}"

if [ -z "${BACKUP_PASSPHRASE:-}" ]; then
    echo "ERROR: BACKUP_PASSPHRASE is not set (needed to decrypt the dump)." >&2
    exit 1
fi

echo "[restore] WARNING: this will overwrite database '$DB_NAME' in '$PG_CONTAINER'."
printf "[restore] Type 'yes' to continue: "
read -r CONFIRM
[ "$CONFIRM" = "yes" ] || { echo "[restore] Aborted."; exit 1; }

# The app keeps writing while the dump loads; stop it and start it again at
# the end, whatever happens.
if [ "$(docker inspect -f '{{.State.Running}}' "$APP_CONTAINER" 2>/dev/null || true)" = "true" ]; then
    echo "[restore] Stopping '$APP_CONTAINER' during the restore..."
    docker stop "$APP_CONTAINER" >/dev/null
    trap 'echo "[restore] Starting '"'$APP_CONTAINER'"' again..."; docker start "$APP_CONTAINER" >/dev/null' EXIT
fi

echo "[restore] Decrypting $FILE and loading into '$DB_NAME'..."

# gpg decrypt (host) -> psql (inside container). Passphrase on fd 3, not in argv
# (visible to other users in ps). One transaction that stops at the first
# error: the dump starts with DROPs (--clean), so a failure halfway used to
# leave tables dropped and still print "Done".
gpg --batch --yes --pinentry-mode loopback --decrypt --passphrase-fd 3 "$FILE" 3<<<"$BACKUP_PASSPHRASE" \
  | docker exec -i "$PG_CONTAINER" psql -U "$DB_USER" -d "$DB_NAME" -v ON_ERROR_STOP=1 --single-transaction -q

echo "[restore] Done."
