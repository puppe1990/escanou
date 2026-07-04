#!/usr/bin/env bash
# Stop mercado dev processes and free the configured port.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

if [[ -f .env ]]; then
  set -a
  # shellcheck disable=SC1091
  source .env
  set +a
fi

PORT="${PORT:-:8090}"
PORT_NUM="${PORT#:}"

echo "=> Parando mercado (porta ${PORT_NUM})…"

pkill -f "${ROOT}/tmp/main" 2>/dev/null || true
pkill -f "${ROOT}.*air -c .air.toml" 2>/dev/null || true
pkill -f "tailwindcss -i ${ROOT}/input.css" 2>/dev/null || true

if command -v lsof >/dev/null 2>&1; then
  PIDS="$(lsof -ti ":${PORT_NUM}" 2>/dev/null || true)"
  if [[ -n "${PIDS}" ]]; then
    echo "=> Liberando :${PORT_NUM}…"
    # shellcheck disable=SC2086
    kill -9 ${PIDS} 2>/dev/null || true
  fi
fi

echo "=> Pronto."