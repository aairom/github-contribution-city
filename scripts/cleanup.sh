#!/usr/bin/env bash

# Resolve script and workspace directories
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
WORKSPACE_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
cd "${WORKSPACE_ROOT}"

echo "🧹 Cleaning up project artifacts, temporary files, and binaries..."

# Stop server if running
if [ -f "${WORKSPACE_ROOT}/scripts/stop.sh" ]; then
  bash "${WORKSPACE_ROOT}/scripts/stop.sh" > /dev/null 2>&1 || true
fi

# Remove build artifacts
rm -rf "${WORKSPACE_ROOT}/bin"
rm -rf "${WORKSPACE_ROOT}/dist"
rm -f "${WORKSPACE_ROOT}/.server.pid"
rm -f "${WORKSPACE_ROOT}/server.log"

# Remove potential node_modules, .venv, __pycache__ if created
rm -rf "${WORKSPACE_ROOT}/node_modules"
rm -rf "${WORKSPACE_ROOT}/.venv"
find "${WORKSPACE_ROOT}" -type d -name "__pycache__" -exec rm -rf {} + 2>/dev/null || true

# Remove generated output files (keep .gitkeep)
find "${WORKSPACE_ROOT}/output" -type f ! -name ".gitkeep" -exec rm -f {} + 2>/dev/null || true

echo "✨ Cleanup complete!"
