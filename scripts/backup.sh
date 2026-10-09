#!/usr/bin/env bash
# ============================================================================
# finance.sh — encrypted Postgres backup
#
# Dumps the finance.sh database (via the running `finance-sh-postgres` container) and
# pipes the plaintext through `gpg --symmetric --cipher-algo AES256` so the
# backup file is encrypted AT REST. Old backups are pruned by BACKUP_RETENTION_DAYS.
#
# COBERTURA: anexos de comprovante vivem em BYTEA no Postgres, então o `pg_dump`
# aqui cobre TUDO — schemas, dados financeiros E os anexos binários. Não há
# object storage externo para fazer backup à parte.
#
# Usage:
#   ./scripts/backup.sh                 # reads .env for creds + passphrase
#   BACKUP_DIR=/mnt/safe ./scripts/backup.sh
#   make backup
#
# Required (from .env or env):
#   DB_USER, DB_NAME                    Postgres role + database
#   BACKUP_PASSPHRASE                   symmetric passphrase for GPG (NOT optional)
# Optional:
#   BACKUP_DIR     (default ./backups)  where encrypted dumps are stored
#   BACKUP_RETENTION_DAYS (default 90)  delete *.sql.gpg older than this
#   PG_CONTAINER   (default finance-sh-postgres)
#
# SECURITY: store BACKUP_DIR on an ENCRYPTED disk (LUKS / encrypted cloud
# volume). The GPG layer protects the file; the encrypted disk protects
# everything else (WAL, temp). See docs/SECURITY.md.
# ============================================================================
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

# Only the keys this script needs, read without executing .env (see env.sh).
# shellcheck source=scripts/env.sh
. "$SCRIPT_DIR/env.sh"
load_env "$ROOT_DIR/.env" DB_USER DB_NAME PG_CONTAINER BACKUP_DIR BACKUP_PASSPHRASE BACKUP_RETENTION_DAYS

DB_USER="${DB_USER:-finance_sh}"
DB_NAME="${DB_NAME:-finance_sh}"
PG_CONTAINER="${PG_CONTAINER:-finance-sh-postgres}"
BACKUP_DIR="${BACKUP_DIR:-$ROOT_DIR/backups}"
# Its own variable: RETENTION_DAYS is the app's LGPD purge window.
RETENTION_DAYS="${BACKUP_RETENTION_DAYS:-90}"

if [ -z "${BACKUP_PASSPHRASE:-}" ]; then
    echo "ERROR: BACKUP_PASSPHRASE is not set (needed to encrypt the dump)." >&2
    echo "       export BACKUP_PASSPHRASE='...'  (keep it out of shell history)" >&2
    exit 1
fi

mkdir -p "$BACKUP_DIR"
TS="$(date +%Y%m%d-%H%M%S)"
OUT="$BACKUP_DIR/finance_sh-${DB_NAME}-${TS}.sql.gpg"

echo "[backup] Dumping '$DB_NAME' from container '$PG_CONTAINER' -> $OUT"

# pg_dump (inside the container) -> gpg symmetric AES256 (on the host) -> file.
# The passphrase goes in on file descriptor 3, not on the command line, where
# any local user could read it from ps or /proc/<pid>/cmdline.
# Written to a .part file and renamed only when the whole pipeline succeeded:
# a pg_dump that died halfway used to leave a truncated dump that looked like
# the newest good backup.
PART="$OUT.part"
trap 'rm -f "$PART"' EXIT
docker exec -i "$PG_CONTAINER" pg_dump -U "$DB_USER" -d "$DB_NAME" --clean --if-exists \
  | gpg --batch --yes --pinentry-mode loopback --symmetric --cipher-algo AES256 \
        --passphrase-fd 3 -o "$PART" 3<<<"$BACKUP_PASSPHRASE"
mv "$PART" "$OUT"
trap - EXIT

echo "[backup] Wrote $(du -h "$OUT" | cut -f1) -> $OUT"

# Retention: prune encrypted dumps older than RETENTION_DAYS.
echo "[backup] Pruning backups older than ${RETENTION_DAYS} days in $BACKUP_DIR"
find "$BACKUP_DIR" -maxdepth 1 -name 'finance_sh-*.sql.gpg' -type f -mtime +"$RETENTION_DAYS" -print -delete || true

echo "[backup] Done."
