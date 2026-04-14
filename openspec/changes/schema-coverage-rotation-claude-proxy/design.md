## Context

The `schema-coverage-rotation` workflow is already structured around deterministic issue-slot gating, repo-memory bookkeeping, repository-local Go commands, and safe outputs that create and assign actionable issues. Today the workflow still uses the `copilot` engine directly.

This proposal is intentionally narrow: it changes only the schema-coverage worker's model-execution path. Unlike the `openspec-verify-label` migration, this workflow should switch engines outright to `claude`, and it should route Claude traffic through the Anthropic-compatible LiteLLM endpoint at `https://elastic.litellm-prod.ai/` using `ANTHROPIC_BASE_URL` and `ANTHROPIC_API_KEY`.

The workflow also runs repository-local commands such as `make setup` and `go run ./scripts/schema-coverage-rotation ...`. Under GH AW, the Claude engine has a shorter default per-tool timeout than Copilot, so the migration should make the execution budget explicit instead of relying on engine defaults.

## Goals / Non-Goals

**Goals:**
- Move the `schema-coverage-rotation` worker to `engine.id: claude`.
- Route Claude requests through the Elastic LiteLLM Anthropic-compatible endpoint via `ANTHROPIC_BASE_URL`.
- Source `ANTHROPIC_API_KEY` from a GitHub Actions secret-backed expression.
- Extend the workflow firewall contract to allow `elastic.litellm-prod.ai`.
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

### 3. Extend the existing firewall contract without broadening it further

The workflow should add `elastic.litellm-prod.ai` to `network.allowed` while keeping the existing `defaults`, `node`, and `go` entries.

Why:
- The workflow already declares an explicit AWF allowlist.
- Claude traffic routed through LiteLLM will fail unless the proxy host is permitted.
- Adding only the required host preserves the current least-privilege posture.

Alternative considered:
- Use a broader or implicit network allowance: rejected because it weakens auditability and expands access beyond what this migration needs.

### 4. Keep downstream issue-assignment behavior unchanged

This change should affect only the schema-coverage analysis worker. The safe output that assigns newly created issues to `acceptance-test-improver` can remain unchanged.

Why:
- The user request is limited to the rotation workflow engine path.
- The assignment target is a separate automation decision and does not need to move in lockstep with the rotation worker.

Alternative considered:
- Switch `assign-to-agent` to a different engine in the same change: rejected because it adds an unrelated behavioral change and complicates rollout.

## Risks / Trade-offs

- [LiteLLM endpoint availability becomes part of schema-coverage rotation reliability] -> Mitigation: make the dependency explicit in workflow frontmatter and specs, then validate on a representative run before rollout.
- [The workflow gains a new provider-auth prerequisite in `ANTHROPIC_API_KEY`] -> Mitigation: require the key to be secret-backed and document it in the workflow contract.
- [Claude may behave differently from Copilot when following the schema-coverage skill] -> Mitigation: keep the prompt, gating, and safe outputs stable and validate with a representative workflow execution.
- [The workflow engine and the downstream assignment agent will differ] -> Mitigation: document that the change is intentionally scoped to the rotation worker only.

## Migration Plan

1. Update `.github/workflows-src/schema-coverage-rotation/workflow.md.tmpl` to switch to `engine.id: claude`, set `ANTHROPIC_BASE_URL`, source `ANTHROPIC_API_KEY` from secrets, add `tools.timeout: 300`, and allow `elastic.litellm-prod.ai`.
2. Regenerate `.github/workflows/schema-coverage-rotation.md` and `.github/workflows/schema-coverage-rotation.lock.yml`.
3. Update workflow generation tests and sync the canonical schema-coverage OpenSpec requirements with the approved engine and firewall contract.
4. Validate the change and run or inspect a representative schema-coverage rotation workflow execution using the Claude + LiteLLM configuration.

## Open Questions

- None for the requested proposal. If the repository later wants a pinned Claude model rather than the engine default, that can be handled as an implementation detail or a follow-up requirements change.
