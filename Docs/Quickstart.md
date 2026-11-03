# Quickstart Guide

This guide walks you through building, configuring, and running **GitHub Contribution City**.

---

## 1. Prerequisites

- **Go**: Version 1.22+ (tested with Go 1.25+)
- Optional: **Podman** or **Docker** for containerized execution.

---

## 2. Configuration

Create a `.env` file from the provided example:

```bash
cp .env.example .env
```

Edit `.env` (optional, for accessing private contributions and higher API rate limits):
```env
GITHUB_TOKEN=ghp_yourPersonalAccessTokenHere
GITHUB_USERNAME=torvalds
CITY_WEEKS=53
CITY_MAX_HEIGHT=14
CITY_THEME=github-dark
PORT=8088
```

---

## 3. Running via Automation Scripts

### Launch Web Server (Detached Mode)
```bash
./scripts/run.sh
```
This builds the binary and launches the application on port `8088` (or custom `$PORT`). It outputs the local URLs:
- **Interactive 3D Web Viewer**: `http://localhost:8088/?user=torvalds&theme=github-dark`
- **SVG Skyline**: `http://localhost:8088/api/svg?user=torvalds&theme=github-dark`
- **3D Isometric SVG**: `http://localhost:8088/api/iso?user=torvalds&theme=cyberpunk`

### Stop Web Server
```bash
./scripts/stop.sh
```

### Cleanup Project Artifacts
```bash
./scripts/cleanup.sh
```

---

## 4. Running via CLI Directly

### Generate SVG Skyline
```bash
go run cmd/city/main.go -user torvalds -format svg -theme github-dark
```

### Generate 3D Isometric SVG
```bash
go run cmd/city/main.go -user torvalds -format iso -theme cyberpunk
```

### Generate All Formats (SVG, 3D ISO, Interactive HTML, JSON)
```bash
go run cmd/city/main.go -user torvalds -format all -weeks 53 -height 14
```
All outputs are saved to the timestamped files in the `output/` directory (e.g. `output/torvalds_city_20261103_074000.svg`).

### Batch Processing
Place text or JSON files containing GitHub usernames in the `input/` folder:
```bash
echo -e "torvalds\nrobpike\nrsc" > input/users.txt
go run cmd/city/main.go -batch input/
```

---

## 5. Running with Podman / Docker

```bash
# Build container image
podman build -t github-contribution-city .

# Run container
podman run -d -p 8088:8088 --name contribution-city github-contribution-city

# Access in browser
open http://localhost:8088/
```
