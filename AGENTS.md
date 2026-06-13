# Antigravity Quota Agent Guide

This document describes `antigravity-quota` specific conventions and tooling.

---

## Project Overview

`antigravity-quota` is a lightweight Go CLI tool that connects to the active Antigravity language server instance to query and print the user's model quota summary.

---

## Common Patterns

```bash
make all          # Full CI pipeline (tidy → build → lint → test)
make test         # Run all tests
make lint         # Run golangci-lint via isolated tool modfile
make tidy         # go mod tidy
make install      # Install to GOBIN
make tools-update # Update tool dependencies
make tools-tidy   # Tidy tool modules
make upgrade-direct-dependencies # Upgrade direct dependencies
```

**Tool Isolation & Version Pinning**: Tools live in `tools/<tool>/go.mod`, invoked as:
```bash
go tool -modfile=tools/golangci-lint/go.mod golangci-lint run ./...
```

To prevent `go mod tidy` from stripping tool dependencies, each tool module contains a `dummy.go` file with a blank import of the tool's main package:
```go
package tools
import _ "github.com/golangci/golangci-lint/v2/cmd/golangci-lint"
```

**CI Gates**: `tidy` must produce no changes (enforced via `git diff --exit-code`).

**Commit Format**:
```
feat[(<topic>)]: <description>
chore[(<topic>)]: <description>
bugfix[(<topic>)]: <description>
```
- Active present tense, one heading only
- Use `*` for bullet points
- Escape backticks in shell

---

## Architecture & Layout

| Component | Directory | Description |
|-----------|-----------|-------------|
| **CLI entry point** | `main.go` | Read environment, fetch data via HTTP POST, format/print output |

---

## Git Workflow

**Branching**: Prefer feature branches for any non-trivial work:
```bash
git checkout -b feature/description
```

**Testing Changes**: Always verify changes using the project's make targets before committing:
```bash
make all          # Run full CI pipeline locally
```

**CRITICAL**: Never push to remote without explicit command. Always verify via `git diff` before pushing.
