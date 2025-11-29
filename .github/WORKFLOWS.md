# GitHub Workflows Documentation

This document describes the GitHub Actions workflows configured for the peat project.

## Workflows Overview

### 1. CI Workflow (`ci.yaml`)

**Triggers:** Push to `main`, Pull Requests to `main`

**Jobs:**

- **Test**
  - Runs on: Linux, macOS, Windows
  - Go versions: 1.24
  - Executes unit tests with race detection
  - Generates code coverage reports

- **Lint**
  - Runs golangci-lint with comprehensive linter configuration
  - Uses the latest version of golangci-lint
  - 5-minute timeout for large codebases

- **Build**
  - Builds the binary using the Makefile
  - Tests that the binary runs correctly
  - Validates basic commands work

### 2. Code Quality Workflow (`quality.yaml`)

**Triggers:** Push to `main`, Pull Requests to `main`

**Jobs:**

- **Format Check**
  - Ensures all Go code is properly formatted with `gofmt`
  - Fails if any files need formatting

- **Go Vet**
  - Runs `go vet` to catch suspicious constructs
  - Checks for common Go programming mistakes

- **Static Check**
  - Runs `staticcheck` for advanced static analysis
  - Catches bugs, performance issues, and style violations

- **Mod Tidy Check**
  - Verifies `go.mod` and `go.sum` are up to date
  - Ensures dependencies are properly tracked

### 3. Security Workflow (`security.yaml`)

**Triggers:** 
- Push to `main`
- Pull Requests to `main`
- Weekly schedule (Mondays at 00:00 UTC)

**Jobs:**

- **Go Vulnerability Check**
  - Uses `govulncheck` to scan for known vulnerabilities
  - Checks both direct and indirect dependencies
  - Runs weekly to catch newly discovered vulnerabilities

- **CodeQL Analysis**
  - Performs semantic code analysis
  - Identifies security vulnerabilities
  - Supports advanced query-based scanning

### 4. Release Workflow (`release.yaml`)

**Triggers:** Git tags (any tag push)

**Jobs:**

- **GoReleaser**
  - Builds binaries for multiple platforms
  - Creates GitHub releases
  - Generates checksums
  - Publishes release artifacts

### 5. Dependabot Auto-merge Workflow (`dependabot-auto-merge.yaml`)

**Triggers:** Pull requests from Dependabot

**Jobs:**

- **Auto-merge Dependabot PRs**
  - Automatically merges patch and minor version updates
  - Requires CI checks to pass
  - Uses squash merge strategy

## Dependabot Configuration

File: `.github/dependabot.yaml`

**Features:**

- **Go Modules Updates**
  - Weekly updates on Mondays at 03:00 UTC
  - Maximum 10 open PRs
  - Labeled with `dependencies` and `go`

- **GitHub Actions Updates**
  - Weekly updates on Mondays at 03:00 UTC
  - Maximum 5 open PRs
  - Labeled with `dependencies` and `github-actions`

## Golangci-lint Configuration

File: `.golangci.yaml`

**Enabled Linters:**

- **Default:** errcheck, gosimple, govet, ineffassign, staticcheck, unused
- **Code Quality:** gofmt, goimports, misspell, revive, stylecheck
- **Security:** gosec
- **Performance:** prealloc, unconvert, unparam, wastedassign
- **Best Practices:** goconst, gocritic, goprintffuncname, nolintlint

**Key Settings:**

- 5-minute timeout
- Local import prefix: `github.com/akasprzok/peat`
- Test files are linted
- Colored output with line numbers

## Required Secrets

The following secrets need to be configured in your GitHub repository:

- `GITHUB_TOKEN` - Automatically provided by GitHub Actions
- `CODECOV_TOKEN` - (Optional) For uploading coverage reports to Codecov

## Status Badges

Add these badges to your README.md:

```markdown
[![CI](https://github.com/akasprzok/peat/workflows/CI/badge.svg)](https://github.com/akasprzok/peat/actions?query=workflow%3ACI)
[![Security](https://github.com/akasprzok/peat/workflows/Security/badge.svg)](https://github.com/akasprzok/peat/actions?query=workflow%3ASecurity)
[![Code Quality](https://github.com/akasprzok/peat/workflows/Code%20Quality/badge.svg)](https://github.com/akasprzok/peat/actions?query=workflow%3A%22Code+Quality%22)
[![Go Report Card](https://goreportcard.com/badge/github.com/akasprzok/peat)](https://goreportcard.com/report/github.com/akasprzok/peat)
```

## Local Development

Before pushing code, you can run the same checks locally:

```bash
# Run all checks
make check

# Individual checks
make fmt      # Format code
make vet      # Run go vet
make lint     # Run golangci-lint
make test     # Run tests

# Build
make build
```

## Workflow Permissions

Each workflow uses the minimum required permissions:

- **CI, Quality:** `contents: read`
- **Security:** `contents: read`, `security-events: write`
- **Release:** `contents: write`
- **Dependabot Auto-merge:** `contents: write`, `pull-requests: write`

## Troubleshooting

### Workflow fails on formatting

Run locally:
```bash
make fmt
git add .
git commit -m "Format code"
```

### Workflow fails on linting

Run locally:
```bash
make lint
```

Fix any reported issues, commit, and push.

### Tests fail in CI but pass locally

- Check if tests are platform-specific
- Verify Go version matches CI
- Check for race conditions (CI runs with `-race`)

### Coverage upload fails

- Ensure `CODECOV_TOKEN` is set in repository secrets
- Check Codecov service status
- The workflow continues even if upload fails (`fail_ci_if_error: false`)

