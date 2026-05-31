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

if [ -f "$ROOT_DIR/.env" ]; then
    set -a
    # shellcheck disable=SC1091
    . "$ROOT_DIR/.env"
    set +a
fi

FILE="${1:-${FILE:-}}"
if [ -z "$FILE" ] || [ ! -f "$FILE" ]; then
    echo "ERROR: pass the encrypted backup file to restore." >&2
    echo "Usage: $0 <path-to.sql.gpg>" >&2
    exit 1
fi

DB_USER="${DB_USER:-finance_sh}"
DB_NAME="${DB_NAME:-finance_sh}"
PG_CONTAINER="${PG_CONTAINER:-finance-sh-postgres}"

if [ -z "${BACKUP_PASSPHRASE:-}" ]; then
    echo "ERROR: BACKUP_PASSPHRASE is not set (needed to decrypt the dump)." >&2
    exit 1
fi

echo "[restore] WARNING: this will overwrite database '$DB_NAME' in '$PG_CONTAINER'."
printf "[restore] Type 'yes' to continue: "
read -r CONFIRM
[ "$CONFIRM" = "yes" ] || { echo "[restore] Aborted."; exit 1; }

echo "[restore] Decrypting $FILE and loading into '$DB_NAME'..."

# gpg decrypt (host) -> psql (inside container).
gpg --batch --yes --decrypt --passphrase "$BACKUP_PASSPHRASE" "$FILE" \
  | docker exec -i "$PG_CONTAINER" psql -U "$DB_USER" -d "$DB_NAME"

echo "[restore] Done."
