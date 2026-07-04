#!/usr/bin/env bash
# One dev entrypoint: stop stale processes, build CSS, tailwind watch + air.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

if [[ -f .env ]]; then
  set -a
  # shellcheck disable=SC1091
  source .env
  set +a
fi

export ENV="${ENV:-development}"
export PORT="${PORT:-:8090}"
export APP_URL="${APP_URL:-http://localhost:${PORT#:}}"
export PORT_STRICT="${PORT_STRICT:-1}"

"${ROOT}/scripts/stop.sh"

AIR=""
if [[ -x "${HOME}/go/bin/air" ]]; then
  AIR="${HOME}/go/bin/air"
elif command -v air >/dev/null 2>&1; then
  AIR="$(command -v air)"
fi

if [[ -z "${AIR}" ]]; then
  echo "air não encontrado. Instale: go install github.com/air-verse/air@latest" >&2
  exit 1
fi

if ! "${AIR}" -v 2>&1 | grep -qi 'air'; then
  echo "O binário 'air' no PATH não é o air-verse (hot reload)." >&2
  echo "Use: go install github.com/air-verse/air@latest" >&2
  echo "Depois: export PATH=\"\${HOME}/go/bin:\${PATH}\"" >&2
  exit 1
fi

echo "=> CSS inicial…"
npx tailwindcss -i input.css -o web/static/css/styles.css

echo "=> Tailwind watch + air em ${APP_URL} (PORT=${PORT})"
npx tailwindcss -i input.css -o web/static/css/styles.css --watch &
TW_PID=$!
cleanup() {
  kill "${TW_PID}" 2>/dev/null || true
}
trap cleanup EXIT INT TERM

exec env ENV="${ENV}" PORT="${PORT}" APP_URL="${APP_URL}" PORT_STRICT="${PORT_STRICT}" \
  "${AIR}" -c .air.toml