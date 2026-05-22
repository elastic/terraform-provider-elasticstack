---
# Go Source Code Analysis Base
# Bundles Serena Go LSP analysis + standard bash permissions for Go source navigation.
#
# Usage:
#   imports:
#     - shared/go-source-analysis.md

imports:
  - shared/serena-go.md

tools:
  bash:
    - "find internal provider -name '*.go' ! -name '*_test.go' -type f"
    - "find internal provider -type f -name '*.go' ! -name '*_test.go'"
    - "find internal/ -maxdepth 1 -ls"
    - "find provider/ -maxdepth 1 -ls"
    - "wc -l internal/**/*.go"
    - "wc -l provider/**/*.go"
    - "head -n * internal/**/*.go"
    - "head -n * provider/**/*.go"
    - "grep -r 'func ' internal provider --include='*.go'"
    - "cat internal/**/*.go"
    - "cat provider/**/*.go"
---

## Go Source Code Analysis Setup

Serena Go LSP analysis is configured for this workspace. Standard bash tools for Go source navigation are available.

### Bash Navigation Tools

Use these bash tools to supplement Serena's semantic analysis:

- `find internal provider -name '*.go' ! -name '*_test.go' -type f` — list all non-test Go source files
- `find internal/ -maxdepth 1 -ls` / `find provider/ -maxdepth 1 -ls` — explore directory structure
- `wc -l internal/**/*.go` / `wc -l provider/**/*.go` — measure file sizes
- `head -n * internal/**/*.go` / `cat internal/**/*.go` — read file contents
- `grep -r 'func ' internal provider --include='*.go'` — find all function definitions
