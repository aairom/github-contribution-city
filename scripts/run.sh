#!/usr/bin/env bash
set -e

# Resolve script and workspace directories
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
WORKSPACE_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
cd "${WORKSPACE_ROOT}"

# Default port
PORT="${PORT:-8088}"
PID_FILE="${WORKSPACE_ROOT}/.server.pid"
LOG_FILE="${WORKSPACE_ROOT}/server.log"

# Check if already running
if [ -f "${PID_FILE}" ]; then
  OLD_PID=$(cat "${PID_FILE}")
  if ps -p "${OLD_PID}" > /dev/null 2>&1; then
    echo "⚠️  Application is already running (PID: ${OLD_PID})."
    echo "🌐 Access URL: http://localhost:${PORT}"
    exit 0
  else
    rm -f "${PID_FILE}"
  fi
fi

# Build binary
echo "🔨 Building GitHub Contribution City..."
go build -o bin/github-contribution-city cmd/city/main.go

# Launch in background / detached mode with nohup
echo "🚀 Starting GitHub Contribution City in background mode on port ${PORT}..."
nohup "${WORKSPACE_ROOT}/bin/github-contribution-city" -server -port "${PORT}" > "${LOG_FILE}" 2>&1 &
SERVER_PID=$!
echo "${SERVER_PID}" > "${PID_FILE}"

# Wait for server ready
sleep 1

# Check if process is still alive
if ps -p "${SERVER_PID}" > /dev/null 2>&1; then
  echo "✅ Application successfully launched (PID: ${SERVER_PID})"
  echo "--------------------------------------------------------"
  echo "🌐 Application URL: http://localhost:${PORT}"
  echo "🌆 Interactive 3D City: http://localhost:${PORT}/?user=torvalds&theme=github-dark"
  echo "📥 SVG Skyline Export: http://localhost:${PORT}/api/svg?user=torvalds"
  echo "📐 3D Isometric SVG:   http://localhost:${PORT}/api/iso?user=torvalds&theme=cyberpunk"
  echo "--------------------------------------------------------"
  echo "📜 Logs: ${LOG_FILE}"
  echo "🛑 To stop: ./scripts/stop.sh"
else
  echo "❌ Failed to start application. Check logs at ${LOG_FILE}:"
  cat "${LOG_FILE}"
  exit 1
fi
