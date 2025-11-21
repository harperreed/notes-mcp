# GitHub Actions Workflows for notes-mcp

**Date:** 2025-11-20
**Status:** Approved for Implementation

## Overview

Automate CI/CD for the Apple Notes MCP server with GitHub Actions workflows for continuous integration, automated releases, and Homebrew distribution.

## Goals

1. Run tests and linting on every PR and push to main
2. Automate releases with GoReleaser when version tags are pushed
3. Publish macOS binaries to GitHub releases
4. Automatically update Homebrew tap formula for easy installation
5. Keep integration tests local-only (documented but not automated)

## Non-Goals

- Automated integration tests in CI (require macOS + Apple Notes setup)
- Cross-platform builds (tool is macOS-specific)
- Docker image publishing (not applicable for this project)
- Self-hosted runners (use GitHub-hosted runners)

## Workflows

### 1. CI Workflow (.github/workflows/ci.yml)

**Triggers:**
- Push to `main` branch
- Pull requests targeting `main`

**Jobs (run in parallel):**

#### Test Job
- **Runner:** ubuntu-latest
- **Steps:**
  1. Checkout code with full history (`fetch-depth: 0`)
  2. Setup Go 1.21+ with caching
  3. Run `go mod download`
  4. Run `make test` (unit tests only, excludes `-tags=integration`)
  5. Run `make build` to verify binary compiles
  6. Upload test coverage to artifacts

#### Lint Job
- **Runner:** ubuntu-latest
- **Steps:**
  1. Checkout code
  2. Setup Go 1.21+ with caching
  3. Run `golangci-lint run` using existing `.golangci.yml` config
  4. Use golangci-lint-action for caching and performance

#### Pre-commit Job
- **Runner:** ubuntu-latest
- **Steps:**
  1. Checkout code
  2. Setup Go 1.21+ with caching
  3. Setup Python for pre-commit framework
  4. Run `pre-commit run --all-files`
  5. Ensures all hooks pass (go fmt, go imports, go-mod-tidy, golangci-lint, go-unit-tests, go build)

**PR Requirements:**
All three jobs must pass before merge is allowed.

### 2. Release Workflow (.github/workflows/release.yml)

**Trigger:**
- Tags matching pattern: `v*` (e.g., `v1.0.0`, `v1.2.3`)

**Permissions:**
- `contents: write` - Create GitHub releases
- `packages: write` - Reserved for future use

**Job: Release**
- **Runner:** ubuntu-latest
- **Steps:**
  1. Checkout with full history (`fetch-depth: 0`)
  2. Setup Go 1.21+ with caching
  3. Run GoReleaser
     - Action: `goreleaser/goreleaser-action@v6`
     - Version: v2 (latest)
     - Command: `release --clean`
     - Token: `HOMEBREW_TAP_TOKEN` (from repository secrets)

**Outputs:**
- GitHub release with changelog
- macOS binaries (darwin/amd64, darwin/arm64)
- Updated Homebrew formula in `harperreed/homebrew-tap`

## GoReleaser Configuration

**.goreleaser.yml:**

```yaml
version: 2

before:
  hooks:
    - go mod tidy
    - go test ./...

builds:
  - main: .
    binary: mcp-apple-notes-go
    goos:
      - darwin
    goarch:
      - amd64
      - arm64
    env:
      - CGO_ENABLED=0
    ldflags:
      - -s -w -X main.version={{.Version}}

archives:
  - format: tar.gz
    name_template: "{{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}"
    files:
      - LICENSE*
      - README*
      - docs/*

brews:
  - repository:
      owner: harperreed
      name: homebrew-tap
      token: "{{ .Env.HOMEBREW_TAP_TOKEN }}"
    folder: Formula
    homepage: "https://github.com/harperreed/notes-mcp"
    description: "MCP server for Apple Notes with CLI tools"
    install: |
      bin.install "mcp-apple-notes-go"
    test: |
      system "#{bin}/mcp-apple-notes-go", "--version"

changelog:
  sort: asc
  filters:
    exclude:
      - '^docs:'
      - '^test:'
      - '^chore:'
```

**Key Features:**
- Builds only for macOS (Intel and Apple Silicon)
- Strips debug symbols (`-s -w`) for smaller binaries
- Injects version via ldflags
- Excludes maintenance commits from changelog
- Tests binary execution in Homebrew formula

## Setup Instructions

### Initial Setup (One-time)

**1. Create Homebrew Tap Repository**
```bash
# On GitHub, create new repository: harperreed/homebrew-tap
# Can be initialized empty - GoReleaser will populate it
```

**2. Generate GitHub Personal Access Token**
- Navigate to: Settings → Developer settings → Personal access tokens → Tokens (classic)
- Click "Generate new token (classic)"
- Token name: `GoReleaser Homebrew Tap`
- Scopes: Select `repo` (all repository permissions)
- Generate and save token securely

**3. Add Token to Repository Secrets**
- Navigate to: notes-mcp repository → Settings → Secrets and variables → Actions
- Click "New repository secret"
- Name: `HOMEBREW_TAP_TOKEN`
- Value: Paste the token from step 2
- Add secret

**4. Install GoReleaser Locally (for testing)**
```bash
brew install goreleaser
```

**5. Test Release Process**
```bash
# From notes-mcp directory
goreleaser release --snapshot --clean

# Verify output in dist/ directory
ls -la dist/
```

### Creating a Release

**1. Prepare Release**
```bash
# Ensure main branch is up to date
git checkout main
git pull

# Create and push tag
git tag v1.0.0
git push origin v1.0.0
```

**2. Monitor Release**
- GitHub Actions automatically triggers release workflow
- Check progress: Actions tab → Release workflow
- Verify GitHub release created with binaries
- Check Homebrew tap updated: `harperreed/homebrew-tap`

**3. Test Installation**
```bash
# Test Homebrew formula
brew install harperreed/tap/notes-mcp

# Verify installation
mcp-apple-notes-go --version
```

### For Contributors

**Running CI Checks Locally:**
```bash
# Run what CI runs
make test          # Unit tests only
make lint          # golangci-lint
make check         # Format + lint + test
pre-commit run --all-files  # All pre-commit hooks

# Full pre-flight check before pushing
make check && pre-commit run --all-files
```

**Integration Tests (Local Only):**

Integration tests require macOS with Apple Notes and are not automated in CI.

```bash
# Ensure Apple Notes is running
open -a "Notes"

# Run integration tests
go test -tags=integration -v ./services -timeout 5m

# Run specific integration test
go test -tags=integration -v ./services -run TestCreateNoteIntegration
```

**Note:** Integration tests create real data in Apple Notes. Manual cleanup may be required.

## Maintenance

### Updating Go Version

**1. Update workflow files:**
```yaml
# .github/workflows/ci.yml and .github/workflows/release.yml
- uses: actions/setup-go@v5
  with:
    go-version: '1.22'  # Update version
```

**2. Update project files:**
- `go.mod`: Update `go 1.22`
- `.tool-versions` (if using): Update go version

**3. Update documentation:**
- README.md: Update requirements

### Updating golangci-lint

**1. Update workflow:**
```yaml
# .github/workflows/ci.yml
- uses: golangci/golangci-lint-action@v4
  with:
    version: v1.55  # Update to match local version
```

**2. Verify locally:**
```bash
golangci-lint --version
make lint
```

### Troubleshooting

**Release fails with Homebrew tap error:**
- Verify `HOMEBREW_TAP_TOKEN` is valid and has `repo` scope
- Check `harperreed/homebrew-tap` repository exists
- Verify token has access to tap repository

**Pre-commit job fails:**
- Run `pre-commit run --all-files` locally
- Fix any formatting or linting issues
- Push fixes

**GoReleaser snapshot fails locally:**
- Ensure all tests pass: `make test`
- Check `.goreleaser.yml` syntax
- Run with verbose logging: `goreleaser release --snapshot --clean --verbose`

## Success Metrics

- CI workflow completes in < 5 minutes
- All PR checks pass before merge
- Releases publish successfully on tag push
- Homebrew formula updates automatically
- Users can install with one command: `brew install harperreed/tap/notes-mcp`

## Future Enhancements (Deferred)

- Automated integration tests on self-hosted macOS runner
- Release notifications (Slack, Discord)
- Dependabot or Renovate for dependency updates
- CodeQL security scanning
- Coverage reporting (Codecov, Coveralls)
