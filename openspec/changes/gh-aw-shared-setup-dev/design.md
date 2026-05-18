## Context

The repository uses GitHub Agentic Workflows (gh-aw v0.72.1) with a compile pipeline: `.github/workflows-src/*/workflow.md.tmpl` → `.github/workflows/*.md` (via Go compiler) → `.github/workflows/*.lock.yml` (via `gh aw compile`).

Workflow templates define both a `pre_activation` job (deterministic gate-checks in native GHA) and an agent job (AI agent steps including setup). Nine workflows repeat the same 4–10 setup steps in their agent phase. Additionally, `code-factory-issue` and `reproducer-factory-issue` embed Elastic Stack setup steps (`make docker-fleet`, `make set-kibana-password`, `make create-es-api-key`, `make setup-kibana-fleet`) that spin up services which are **not accessible from the agent's chroot sandbox** — the agent runs unaware of these services.

The current duplication matrix shows the same four tool installations across the majority of workflows:

| Step | Count |
|------|-------|
| `actions/setup-go@v6` | 7 |
| `actions/setup-node@v6` | 8 |
| `hashicorp/setup-terraform@v4` | 3 |
| `Export Go paths for AWF chroot mode` | 7 |
| `make setup` or `npm ci` | 9 |

## Goals / Non-Goals

**Goals:**
- Replace all per-workflow agent-phase setup steps with a single GH-AW `imports:` reference
- Remove non-functional Elastic Stack setup scaffolding from factory-issue workflows
- Ensure Terraform PATH is always exported into chroot when Terraform CLI is installed (fixes latent inconsistency)
- Keep the `pre_activation` checkout/Go steps in `ci-deadcode-removal-rotation` untouched (they run before the agent and have specific needs)

**Non-Goals:**
- No changes to `pre_activation` job logic or gate-checks
- No changes to agent prompts, engine config, or safe-outputs
- No new Elastic Stack accessibility work (removing broken scaffolding explicitly declared out of scope)
- No changes to the workflow compiler (`scripts/compile-workflow-sources/`)

## Decisions

### 1. Zero-option shared import (no `import-schema`)
**Rationale**: Every workflow that needs repo tooling needs Go, Terraform, Node, and `make setup`. Adding boolean flags (`setup-go: true`, `setup-terraform: false`) creates decision fatigue and brittle configurations. If a workflow doesn't need a tool, the step runs in seconds — negligible cost for zero cognitive overhead. This mirrors how monorepo CI templates work.

### 2. Terraform PATH export inside the Go chroot step, gated by `which terraform`
**Rationale**: `code-factory-issue` previously exported `TERRAFORM_BIN` and updated `PATH` in its chroot step, while `reproducer-factory-issue` did not despite installing Terraform. The shared workflow installs Terraform before the chroot export, then conditionally appends Terraform paths using a shell `if [ -x "$(which terraform)" ]` check. This is universally correct: if Terraform is there, expose it; if not, harmless no-op.

### 3. `make setup` subsumes `npm ci`
**Rationale**: `change-factory-issue` and `research-factory-issue` ran `npm ci` directly. `make setup` includes `setup-openspec` which runs `npm ci`. Replacing `npm ci` with `make setup` is consistent and adds negligible overhead (Go tooling already installed by the shared step anyway).

### 4. `shared/setup-dev.yml` lives inside `.github/workflows/shared/`
**Rationale**: GH-AW resolves `shared/setup-dev.yml` from `.github/workflows/` when imported from `.github/workflows-src/*/workflow.md.tmpl`. The `shared/` convention aligns with GH-AW documentation examples. The file has no `on:` field so it is validated but never compiled into a standalone workflow.

### 5. Remove Elastic Stack steps entirely
**Rationale**: The steps (`make docker-fleet`, `make set-kibana-password`, etc.) assume a Docker Compose environment where the agent can reach Elasticsearch/Kibana. In practice, agents run in a chroot with isolated networking. Removing them eliminates misleading infrastructure setup that the agent cannot use. If Elastic Stack access is needed later, it should be designed as a separate shared import that sets up network-accessible services correctly.

## Risks / Trade-offs

| Risk | Mitigation |
|------|-----------|
| `make setup` on agents that previously only ran `npm ci` does slightly more work (Go fmt tools, vendor check) | Steps run in seconds; negligible cost |
| `ci-deadcode-removal-rotation` now runs `actions/setup-go@v6` twice (once in `pre_activation`, once via shared import in agent phase) | Harmless duplication; `cache: false` means no cache contention |
| Imported steps run before any workflow-specific agent steps | Correct — setup must precede agent logic |
| GH-AW compile-time injection of imported steps could produce YAML ordering issues with step-level `if:` conditions | No imported steps have `if:` conditions; all unconditional |
| Regenerating `.lock.yml` files requires valid `gh aw` CLI access | CI will compile on push; locally, `make compile-workflows` handles the first hop |

## Migration Plan

1. Create `.github/workflows/shared/setup-dev.yml`
2. Modify each `.md.tmpl`:
   - Add `imports: [shared/setup-dev.yml]` to frontmatter
   - Delete all agent-phase setup steps (Go, TF, Node, chroot export, make setup, npm ci, Elastic Stack)
3. Run `make compile-workflows` to regenerate `.md` files
4. Run `gh aw compile` (or let CI do it) to regenerate `.lock.yml` files
5. Verify `make check-lint` passes on compiled outputs

## Open Questions

- None outstanding.
