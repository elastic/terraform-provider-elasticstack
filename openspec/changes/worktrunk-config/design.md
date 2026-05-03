## Context

The repo is a bare git repository at `terraform-provider-elasticstack/` with worktrees checked out per branch. A `worktrunk` worktree exists as a sibling outside the bare repo; feature-branch worktrees live inside it (e.g. `terraform-provider-elasticstack/ml-start-time`).

Worktrunk is installed via Homebrew but has no user or project config. Docker Compose uses hardcoded container names and ports, so only one worktree can run a stack at a time. Developers currently maintain `.env` manually.

## Goals / Non-Goals

**Goals:**
- Make new worktrees self-configuring: `make setup` runs on creation; `.env` is generated with unique ports
- Isolate docker stacks per worktree so multiple stacks can run in parallel without conflict
- Run `make check-lint` automatically before every commit
- Tear down a worktree's docker stack automatically when the worktree is removed
- Fix Makefile targets that hardcode port 9200/5601 so they work in any worktree

**Non-Goals:**
- Auto-starting docker stacks on worktree creation (manual `docker compose up` is preferred)
- TLS docker-compose configuration (`docker-compose.tls.yml`) — ports not currently used in parallel workflows
- Changing the acceptance test invocation flow

## Decisions

### Worktree path template: inside bare repo

`worktree-path = "{{ repo_path }}/{{ branch | sanitize }}"`

This matches the existing layout (`terraform-provider-elasticstack/ml-start-time` etc.) and keeps all worktrees in one place. The `worktrunk` worktree is a legacy exception at the sibling level; new worktrees follow the inside-bare pattern.

Alternatives considered:
- Sibling directories (`../repo.branch`) — inconsistent with existing layout, more filesystem noise at the project level.

### Docker isolation: remove container_name, use hash_port for ports

Removing all `container_name:` directives from `docker-compose.yml` lets Docker Compose namespace containers and volumes automatically by project name (the worktree directory name). This is zero-config isolation for containers and volumes.

Ports require explicit unique values since they're host-bound. The `post-start` hook generates `.env` with:
- `ELASTICSEARCH_PORT={{ branch | hash_port }}` — port in range 10000–19999
- `KIBANA_PORT={{ (branch ~ '-kb') | hash_port }}` — different range slot for same branch

Alternatives considered:
- Explicit port offsets (e.g. `9200 + N`) — requires a registry to avoid collisions.
- Docker-internal networking only (no host ports) — breaks `make testacc` which connects from the host.

### .env generation: .env.template + post-start appends dynamic lines

`.env.template` is committed and holds all static configuration (stack version, passwords, Go version, etc.). The `post-start` hook copies it to `.env` and appends the three port-dependent lines.

This separates concerns cleanly: static config is versioned alongside the code; per-worktree values are generated at worktree creation time and never committed.

Alternatives considered:
- Generate `.env` entirely from the hook — static values buried in `.config/wt.toml`, harder to update `STACK_VERSION` at release time.
- `wt step copy-ignored` — copies `.env` from source worktree unchanged; ports would not be unique.

### pre-remove for docker teardown

`pre-remove` runs while the worktree directory still exists, so `docker compose down --volumes` can derive the correct Compose project name from `cwd`. `post-remove` runs in the primary worktree (removed dir is gone) and would tear down the wrong stack.

### Makefile ports: `?=` defaults

Add `ELASTICSEARCH_PORT ?= 9200` and `KIBANA_PORT ?= 5601` near the top of the Makefile. Swap hardcoded literals to `$(ELASTICSEARCH_PORT)` / `$(KIBANA_PORT)` in affected targets. The `?=` means the Makefile is backwards-compatible when run outside a worktree without `.env` loaded.

Developers running these targets inside a worktree need to export `.env` first (e.g. `set -a; . ./.env; set +a`) or pass the vars explicitly. This is not automated — a note in dev-docs is sufficient.

## Risks / Trade-offs

- **hash_port collisions** → Low probability with ≤10 worktrees across a 10 000-port range; mitigated by using distinct hash inputs for ES and KB ports per branch.
- **`make check-lint` runs `make setup` on every pre-commit** → Setup targets are idempotent (stamp files, vendor cache) so subsequent runs are fast; first run after a clean clone may add ~30s. Acceptable given lint correctness is the goal.
- **Developers must export `.env` before using Makefile port variables directly** → Documented in dev-docs; `docker compose` commands (which auto-load `.env`) are unaffected.

## Migration Plan

1. Update `docker-compose.yml` — remove `container_name:` directives
2. Add `.env.template` — copy current `.env`, strip container name vars, strip port vars
3. Add `.config/wt.toml` with hooks
4. Update Makefile — add `?=` defaults, fix hardcoded literals
5. Update dev-docs with user config instructions (worktree path template, shell integration)
6. For existing worktrees: run `post-start` hook manually (`wt hook post-start`) to generate a fresh `.env` with unique ports, or generate manually

No rollback needed — all changes are additive or cosmetic in docker-compose.yml.
