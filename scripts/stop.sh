#!/usr/bin/env bash

# Resolve script and workspace directories
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
WORKSPACE_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
cd "${WORKSPACE_ROOT}"

echo "🛑 Stopping GitHub Contribution City server..."

PIDS=$(pgrep -f "bin/github-contribution-city" || true)
if [ -n "${PIDS}" ]; then
  kill ${PIDS} 2>/dev/null || true
  sleep 1
  echo "✅ Application stopped (PIDs: ${PIDS})."
else
  echo "ℹ️  No running instance found."
fi

rm -f "${WORKSPACE_ROOT}/.server.pid"
