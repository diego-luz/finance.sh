#!/usr/bin/env bash
#
# Creates .env from .env.example with secrets generated for THIS installation.
#
# The repository deliberately ships no working secret: a value committed here is
# a value every reader of the project already has. This script fills the blanks
# with random ones so a fresh deploy is never protected by a published key.
#
#   ./scripts/gen-env.sh
#
# It refuses to touch an existing .env. Rotating secrets in place is not a
# generation problem: JWT secrets can be replaced freely (everyone is logged
# out), DB_PASSWORD must also be changed inside Postgres (ALTER ROLE), and
# ENCRYPTION_KEY cannot be swapped without re-encrypting stored data.
set -euo pipefail

cd "$(dirname "$0")/.."

readonly EXAMPLE=.env.example
readonly TARGET=.env

if [[ ! -f $EXAMPLE ]]; then
  echo "erro: $EXAMPLE não encontrado (rode a partir do repositório)" >&2
  exit 1
fi

if [[ -e $TARGET ]]; then
  echo "erro: $TARGET já existe — não vou sobrescrever." >&2
  echo "      Para começar do zero: mv $TARGET $TARGET.bak && $0" >&2
  exit 1
fi

if ! command -v openssl >/dev/null 2>&1; then
  echo "erro: openssl não encontrado (necessário para gerar os segredos)" >&2
  exit 1
fi

# 32 random bytes, base64 — the shape ENCRYPTION_KEY requires (AES-256) and a
# fine size for the JWT signing secrets too.
secret() { openssl rand -base64 32; }
# Password without base64 padding characters, which would need quoting in a URL.
password() { openssl rand -hex 24; }

ENCRYPTION_KEY=$(secret)
JWT_ACCESS_SECRET=$(secret)
JWT_REFRESH_SECRET=$(secret)
DB_PASSWORD=$(password)

# Fill each placeholder with its generated value; every other line is copied
# through untouched, so comments and defaults stay as documented.
# The secrets reach awk through the environment, not as -v arguments that any
# local user could read in ps; and the file is born 0600 (umask), instead of
# 0644 until the chmod below.
export ENCRYPTION_KEY JWT_ACCESS_SECRET JWT_REFRESH_SECRET DB_PASSWORD
(
  umask 077
  awk '
    /^ENCRYPTION_KEY=/     { print "ENCRYPTION_KEY=" ENVIRON["ENCRYPTION_KEY"];         next }
    /^JWT_ACCESS_SECRET=/  { print "JWT_ACCESS_SECRET=" ENVIRON["JWT_ACCESS_SECRET"];   next }
    /^JWT_REFRESH_SECRET=/ { print "JWT_REFRESH_SECRET=" ENVIRON["JWT_REFRESH_SECRET"]; next }
    /^DB_PASSWORD=/        { print "DB_PASSWORD=" ENVIRON["DB_PASSWORD"];               next }
    { print }
  ' "$EXAMPLE" > "$TARGET"
)

chmod 600 "$TARGET"

echo "$TARGET criado com segredos próprios desta instalação (modo 600)."
echo
echo "  ENCRYPTION_KEY      gerada  <- guarde num cofre: sem ela os dados"
echo "                              criptografados ficam ilegíveis"
echo "  JWT_ACCESS_SECRET   gerada"
echo "  JWT_REFRESH_SECRET  gerada"
echo "  DB_PASSWORD         gerada  <- aplicada ao criar o volume do Postgres"
echo
echo "Revise o restante (APP_ENV, FRONTEND_URL, SMTP) e suba com: docker compose up -d"
