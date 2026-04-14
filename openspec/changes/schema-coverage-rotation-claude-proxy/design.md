## Context

The `schema-coverage-rotation` workflow is already structured around deterministic issue-slot gating, repo-memory bookkeeping, repository-local Go commands, and safe outputs that create and assign actionable issues. **Before this change**, the rotation worker ran on the GitHub-hosted **`copilot`** engine.

This proposal is intentionally narrow: it changes only the schema-coverage worker's model-execution path. Unlike the `openspec-verify-label` migration, **this change** switches the worker engine outright to **`claude`** and routes Claude traffic through the Anthropic-compatible LiteLLM endpoint at `https://elastic.litellm-prod.ai/` using `ANTHROPIC_BASE_URL` and `ANTHROPIC_API_KEY`.

The workflow also runs repository-local commands such as `make setup` and `go run ./scripts/schema-coverage-rotation ...`. Under GH AW, the Claude engine has a shorter default per-tool timeout than Copilot, so the migration should make the execution budget explicit instead of relying on engine defaults.

## Goals / Non-Goals

**Goals:**
- Move the `schema-coverage-rotation` worker to `engine.id: claude`.
- Route Claude requests through the Elastic LiteLLM Anthropic-compatible endpoint via `ANTHROPIC_BASE_URL`.
- Source `ANTHROPIC_API_KEY` from a GitHub Actions secret-backed expression.
- Extend the authored workflow `network.allowed` with `elastic.litellm-prod.ai` and accept the compiled Claude lock’s broader AWF domain bundle (see Decision 3).
- Preserve the workflow's existing issue-slot gating, repo-memory flow, issue creation rules, and `assign-to-agent` behavior.
- Set an explicit Claude-compatible tool timeout for the workflow's repository-local commands.

**Non-Goals:**
- Changing how entities are selected, analyzed, or recorded in repo-memory.
- Changing the issue content rubric or the `acceptance-test-improver` assignment flow.
- Migrating other GH AW workflows in the same change.
- Introducing a separate custom Claude model pin unless the repository later decides it needs one.

## Decisions

### 1. Switch the workflow engine to Claude and route it with Anthropic-compatible env vars

The workflow should move from `engine.id: copilot` to `engine.id: claude` and configure:

- `engine.env.ANTHROPIC_BASE_URL: https://elastic.litellm-prod.ai/`
- `engine.env.ANTHROPIC_API_KEY` sourced from a GitHub Actions secret expression

Why:
- The requested migration explicitly says to use Claude as the engine.
- GH AW documents `ANTHROPIC_BASE_URL` as the supported path for routing Claude through a custom Anthropic-compatible endpoint.
- Secret-backed `ANTHROPIC_API_KEY` keeps provider auth out of the repository while making the operational requirement explicit.

Alternatives considered:
- Keep the workflow on Copilot and use the same BYOK pattern as `openspec-verify-label`: rejected because this workflow request explicitly switches to Claude.
- Use a broader engine endpoint override without `ANTHROPIC_BASE_URL`: rejected because the request specifically names the Anthropic env var path and GH AW treats it as a first-class routing mechanism.

### 2. Set an explicit Claude tool timeout

The workflow should set `tools.timeout: 300` so repository-local commands have a stable per-tool-call budget after the engine switch.

Why:
- Claude's GH AW default per-tool timeout is shorter than Copilot's effective behavior.
- The workflow runs `make setup` and `go run` commands that may exceed Claude's default 60-second budget.
- Encoding the timeout in the workflow contract makes the migration behavior explicit and testable.

Alternative considered:
- Rely on Claude defaults: rejected because the workflow already depends on longer-running repository-local commands and would become more brittle after the engine switch.

### 3. Author the LiteLLM host in `network.allowed` while accepting compiler-expanded AWF domains in the lock

The authored workflow should add `elastic.litellm-prod.ai` to `network.allowed` while keeping the existing `defaults`, `node`, and `go` entries. That is the maintainers’ explicit contract in YAML.

`gh aw compile` then emits a lock file where the Claude engine’s AWF integration merges those keys with a **fixed, compiler-supplied domain bundle** (toolchains, package registries, Claude runtime hosts, and similar). The practical egress surface at runtime is therefore **larger than the four strings in frontmatter**; it is not “only” adding the LiteLLM hostname. This migration intentionally narrows what **repository authors must declare** (including LiteLLM) while acknowledging the reviewed trade-off that **effective** network scope follows the compiler-generated lock.

Why:
- LiteLLM traffic fails unless `elastic.litellm-prod.ai` appears in the authored allowlist the compiler consumes.
- The Claude engine path does not reduce the lock to authored keys alone; documenting both layers avoids overstating how small the firewall change is.

Alternative considered:
- Treat the authored YAML list as the complete runtime firewall: rejected because it misrepresents how GH AW Claude locks behave today.

### 4. Keep downstream issue-assignment behavior unchanged

This change should affect only the schema-coverage analysis worker. The safe output that assigns newly created issues to `acceptance-test-improver` can remain unchanged.

Why:
- The user request is limited to the rotation workflow engine path.
- The assignment target is a separate automation decision and does not need to move in lockstep with the rotation worker.

Alternative considered:
- Switch `assign-to-agent` to a different engine in the same change: rejected because it adds an unrelated behavioral change and complicates rollout.

## Risks / Trade-offs

- [LiteLLM endpoint availability becomes part of schema-coverage rotation reliability] -> Mitigation: make the dependency explicit in workflow frontmatter and specs, then validate on a representative run before rollout.
- [The workflow gains a new provider-auth prerequisite: the Actions secret `CLAUDE_LITELLM_PROXY_API_KEY` mapped to `ANTHROPIC_API_KEY` at runtime] -> Mitigation: document the secret in change artifacts and require secret-backed configuration only (no literals in-repo).
- [Claude may behave differently from Copilot when following the schema-coverage skill] -> Mitigation: keep the prompt, gating, and safe outputs stable and validate with a representative workflow execution.
- [The workflow engine and the downstream assignment agent will differ] -> Mitigation: document that the change is intentionally scoped to the rotation worker only.

## Operator prerequisites

Operators must configure the GitHub Actions secret **`CLAUDE_LITELLM_PROXY_API_KEY`** for repositories that run this workflow; the workflow sources it as `ANTHROPIC_API_KEY` for the Claude engine.

## Migration Plan

1. Update `.github/workflows-src/schema-coverage-rotation/workflow.md.tmpl` to switch to `engine.id: claude`, set `ANTHROPIC_BASE_URL`, source `ANTHROPIC_API_KEY` from `${{ secrets.CLAUDE_LITELLM_PROXY_API_KEY }}`, add `tools.timeout: 300`, and allow `elastic.litellm-prod.ai` in authored `network.allowed` (then regenerate the lock and review the compiler’s AWF domain bundle as needed).
2. Regenerate `.github/workflows/schema-coverage-rotation.md` and `.github/workflows/schema-coverage-rotation.lock.yml`.
3. Update workflow generation tests and sync the canonical schema-coverage OpenSpec requirements with the approved engine and firewall contract.
4. Validate the change and run or inspect a representative schema-coverage rotation workflow execution using the Claude + LiteLLM configuration.

## Open Questions

- None for the requested proposal. If the repository later wants a pinned Claude model rather than the engine default, that can be handled as an implementation detail or a follow-up requirements change.
