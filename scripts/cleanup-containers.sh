#!/usr/bin/env bash

# Delete container images built for this project
IMAGE_NAME="${1:-github-contribution-city}"

echo "🐳 Cleaning up container images for '${IMAGE_NAME}'..."

if command -v podman &> /dev/null; then
  echo "Using podman..."
  podman rmi -f "${IMAGE_NAME}" 2>/dev/null || true
  podman image prune -f 2>/dev/null || true
elif command -v docker &> /dev/null; then
  echo "Using docker..."
  docker rmi -f "${IMAGE_NAME}" 2>/dev/null || true
  docker image prune -f 2>/dev/null || true
fi

echo "✅ Container cleanup finished."
