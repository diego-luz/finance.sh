# shellcheck shell=bash
# Reads selected keys from the project's .env WITHOUT executing it.
#
# Sourcing .env ran it as shell code: an unquoted value with a space
# (ADMIN_ORG_NAME=Minha Organização, as gen-env.sh writes it) became a command
# and, under set -e, aborted backup/restore on a fresh install. This reads
# KEY=value lines instead: surrounding quotes are removed, and a trailing
# " # comment" is dropped from unquoted values. Variables already set in the
# environment win over the file.
#
#   . "$SCRIPT_DIR/env.sh"; load_env "$ROOT_DIR/.env" DB_USER DB_NAME ...
load_env() {
    local file="$1" key line value
    shift
    [ -f "$file" ] || return 0
    for key in "$@"; do
        [ -n "${!key:-}" ] && continue
        line="$(grep -E "^${key}=" "$file" | tail -n 1 || true)"
        [ -n "$line" ] || continue
        value="${line#*=}"
        case "$value" in
            \"*) value="${value#\"}"; value="${value%%\"*}" ;;
            \'*) value="${value#\'}"; value="${value%%\'*}" ;;
            *)   value="$(printf '%s' "$value" | sed -E 's/[[:space:]]+#.*$//; s/[[:space:]]+$//')" ;;
        esac
        printf -v "$key" '%s' "$value"
        export "${key?}"
    done
}
