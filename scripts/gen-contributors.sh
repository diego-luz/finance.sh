#!/usr/bin/env bash
# gen-contributors.sh — regenera o bloco "Contribuidores de código" do
# CONTRIBUTORS.md a partir de `git shortlog -sn --no-merges`.
#
# Uso:
#   scripts/gen-contributors.sh             # atualiza CONTRIBUTORS.md
#   scripts/gen-contributors.sh --check     # falha se estiver desatualizado (CI)
#   scripts/gen-contributors.sh --print     # só imprime o bloco em stdout
#
# Filtra bots conhecidos (dependabot, github-actions, renovate). Ignora commits
# de merge. Conta commits por (nome, email) e renderiza tabela markdown.
#
# Marcadores no arquivo:
#   <!-- CONTRIBUTORS:START -->
#   ... bloco gerado ...
#   <!-- CONTRIBUTORS:END -->

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
FILE="${ROOT}/CONTRIBUTORS.md"
START_MARK="<!-- CONTRIBUTORS:START -->"
END_MARK="<!-- CONTRIBUTORS:END -->"

# Mantenedores listados separadamente — não duplicar na tabela.
MAINTAINERS_EMAILS=(
  "diego@finance.sh"
  "diegoalvescs@gmail.com"
)

# Bots conhecidos que não entram na lista.
BOT_PATTERNS='dependabot|github-actions|renovate|pre-commit-ci|imgbot|allcontributors'

mode="apply"
case "${1:-}" in
  --check) mode="check" ;;
  --print) mode="print" ;;
  --help|-h)
    sed -n '2,18p' "$0"
    exit 0
    ;;
esac

cd "$ROOT"

if ! git rev-parse --git-dir >/dev/null 2>&1; then
  echo "fatal: not inside a git repository" >&2
  exit 1
fi

if ! git log -1 >/dev/null 2>&1; then
  block="$(cat <<'EOF'
_Lista vazia por enquanto. Será preenchida automaticamente pelo `scripts/gen-contributors.sh` quando houver primeiros commits após a release pública v0.1.0._
EOF
)"
else
  # Build the table from shortlog. Format: "  <N>\t<NAME>\t<EMAIL>"
  raw="$(git log --no-merges --format='%aN%x09%aE' \
    | grep -viE "$BOT_PATTERNS" \
    | sort \
    | uniq -c \
    | sort -k1,1 -rn)"

  # Drop maintainers from the contributors table (avoid double listing).
  filtered="$raw"
  for m in "${MAINTAINERS_EMAILS[@]}"; do
    filtered="$(printf '%s\n' "$filtered" | grep -vF "$m" || true)"
  done

  if [ -z "$filtered" ]; then
    block="$(cat <<'EOF'
_Sem contribuidores externos ainda. Seja o primeiro: veja [CONTRIBUTING.md](CONTRIBUTING.md)._
EOF
)"
  else
    block_header="$(cat <<'EOF'
| Contribuidor | Commits |
|---|---:|
EOF
)"
    block_body="$(printf '%s\n' "$filtered" | awk -F'\t' '
      {
        # Remove leading whitespace + extract commit count.
        gsub(/^[ \t]+/, "", $1)
        split($1, parts, " ")
        commits = parts[1]
        name = parts[2]
        for (i=3; i<=length(parts); i++) name = name " " parts[i]
        email = $2
        # Guess GitHub username from common email patterns.
        gh = ""
        if (match(email, /^[0-9]+\+([^@]+)@users\.noreply\.github\.com$/, m)) gh = m[1]
        else if (match(email, /^([^@]+)@users\.noreply\.github\.com$/, m))    gh = m[1]
        if (gh != "") {
          printf("| [%s](https://github.com/%s) | %s |\n", name, gh, commits)
        } else {
          printf("| %s | %s |\n", name, commits)
        }
      }
    ')"
    block="${block_header}
${block_body}"
  fi
fi

# Render the block content (without markers).
rendered_block="$block"

if [ "$mode" = "print" ]; then
  printf '%s\n' "$rendered_block"
  exit 0
fi

if [ ! -f "$FILE" ]; then
  echo "fatal: $FILE not found" >&2
  exit 1
fi

# Splice the new block between the markers.
new="$(awk -v start="$START_MARK" -v end="$END_MARK" -v block="$rendered_block" '
  BEGIN { in_block = 0; printed = 0 }
  {
    if ($0 ~ start) {
      print
      print ""
      print block
      print ""
      in_block = 1
      printed = 1
      next
    }
    if ($0 ~ end) {
      in_block = 0
      print
      next
    }
    if (!in_block) print
  }
  END {
    if (!printed) {
      print "fatal: markers not found in CONTRIBUTORS.md" > "/dev/stderr"
      exit 2
    }
  }
' "$FILE")"

if [ "$mode" = "check" ]; then
  if ! diff -q <(printf '%s\n' "$new") "$FILE" >/dev/null; then
    echo "CONTRIBUTORS.md is out of date. Run: scripts/gen-contributors.sh" >&2
    diff <(printf '%s\n' "$new") "$FILE" || true
    exit 1
  fi
  echo "CONTRIBUTORS.md is up to date."
  exit 0
fi

printf '%s\n' "$new" > "$FILE"
echo "CONTRIBUTORS.md updated."
