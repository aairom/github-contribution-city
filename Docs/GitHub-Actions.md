# GitHub Actions CI/CD and Automation Guide

This guide details the complete GitHub Actions automation infrastructure implemented for **GitHub Contribution City**.

---

## 1. Overview of Workflows

The repository includes a production-grade CI/CD automation suite located in [`.github/workflows/`](../.github/workflows/):

| Workflow | File | Trigger | Description |
| :--- | :--- | :--- | :--- |
| **Continuous Integration** | [`ci.yml`](../.github/workflows/ci.yml) | Push & PR to `main`/`master` | Formats check, `go vet`, `golangci-lint`, `govulncheck`, multi-OS unit tests (`ubuntu`, `macos`, `windows`), coverage upload, E2E generation validation, and multi-arch cross-compilation. |
| **Container Image Publishing** | [`docker.yml`](../.github/workflows/docker.yml) | Push to `main` & version tags | Builds multi-arch (`linux/amd64`, `linux/arm64`) container images and publishes to GitHub Container Registry (`ghcr.io`). |
| **Daily Visualization Update** | [`update-city.yml`](../.github/workflows/update-city.yml) | Daily Cron (`03:17 UTC`) & `workflow_dispatch` | Fetches fresh contribution calendars for the repository owner and any tracked users in `input/`, regenerates SVGs and HTML, and commits changes back to the repository (`[skip ci]`). |
| **Release Management** | [`release.yml`](../.github/workflows/release.yml) | Tag push (`v*.*.*`) | Automatically cross-compiles release binaries for Linux, macOS, and Windows (AMD64 & ARM64), packages `.tar.gz` and `.zip` archives with SHA256 checksums, and creates a GitHub Release. |
| **GitHub Pages Deployment** | [`docs.yml`](../.github/workflows/docs.yml) | Doc changes & `workflow_dispatch` | Generates interactive 3D WebGL demo city and documentation, deploying it to GitHub Pages. |
| **Dependency Scanning** | [`.github/dependabot.yml`](../.github/dependabot.yml) | Weekly | Scans and generates automated update PRs for Go modules, GitHub Actions, and Docker containers. |

---

## 2. Secrets Configuration & Setup

### Required & Optional Secrets

Go to your repository settings: **Settings > Secrets and variables > Actions**.

| Secret Name | Required? | Description |
| :--- | :--- | :--- |
| `GITHUB_TOKEN` | Built-in | Automatically provided by GitHub Actions with `contents: write` and `packages: write` permissions. |
| `CONTRIBUTION_TOKEN` | Optional | A Personal Access Token (PAT) with `read:user` and `repo` scopes. Supply this if you want the daily automated job to include **private and internal contributions** in the generated skyline. |

---

## 3. GitHub Pages Setup

To activate live documentation & 3D demo hosting:
1. Navigate to **Settings > Pages**.
2. Under **Build and deployment > Source**, select **GitHub Actions**.
3. Push to `main` or trigger the **Deploy Documentation to GitHub Pages** workflow manually.

---

## 4. Tracked Users Batch Processing

To automatically generate and maintain contribution cities for multiple team members or friends:
1. Place a `.txt` or `.json` file in the [`input/`](../input/) folder (e.g. `input/team.txt`):
   ```text
   torvalds
   robpike
   rsc
   ```
2. The daily scheduled job will automatically pick up all files in `input/` recursively and generate individual timestamped SVG skylines and 3D scenes in `output/`.

---

## 5. Local Makefile Commands for CI Emulation

You can run CI checks locally prior to pushing:

```bash
# Run tests with race detection and coverage
make test

# Generate HTML coverage report
make coverage

# Verify lint and code formatting
make lint

# Cross-compile for Linux, macOS, and Windows
make build-all
```
