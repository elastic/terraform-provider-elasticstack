## Why

Before this change, the `schema-coverage-rotation` workflow ran its rotation worker on the GitHub-hosted Copilot engine. This change moves it onto the same LiteLLM routing direction as other GH AW migrations, with the worker using the `claude` engine through an Anthropic-compatible proxy at `https://elastic.litellm-prod.ai/`.

## What Changes

- Update the `schema-coverage-rotation` workflow contract so the rotation worker uses `engine.id: claude`.
- Route Claude API traffic through the Elastic LiteLLM Anthropic-compatible endpoint via `ANTHROPIC_BASE_URL` and a secret-backed `ANTHROPIC_API_KEY`.
- Extend the authored workflow `network.allowed` with `elastic.litellm-prod.ai` alongside the existing bootstrap ecosystems. The compiled Claude lock file still reflects a **broader** compiler-generated AWF domain bundle for agent execution (merged into AWF `--allow-domains`), not only the keys declared in YAML; this change accepts that reviewed trade-off.
- Require an explicit tool timeout suitable for the Claude engine so repository-local schema analysis commands are not constrained by Claude's shorter default tool-call budget.
- Preserve the existing issue-slot gating, repo-memory flow, issue creation rules, and downstream `assign-to-agent` behavior.

## Operator prerequisites

This workflow depends on the GitHub Actions repository secret **`CLAUDE_LITELLM_PROXY_API_KEY`**. Configure it in the repo (or organization) secrets used by Actions before relying on scheduled or manual runs.

## Capabilities

### Modified Capabilities

- `ci-schema-coverage-rotation-toolchain`: Extend the existing canonical spec so it also covers the Claude engine, Anthropic-compatible LiteLLM proxy configuration, explicit per-tool timeout, and the LiteLLM host in authored `network.allowed` (structural allowlist pairing stays on this single capability id).

## Impact

- Authored workflow source under `.github/workflows-src/schema-coverage-rotation/`
- Generated workflow artifacts under `.github/workflows/schema-coverage-rotation.*`
- GH AW engine authentication, new secret **`CLAUDE_LITELLM_PROXY_API_KEY`**, and effective AWF network surface (authored allowlist keys plus compiler-expanded domain set in the lock)
- Workflow generation tests under `.github/workflows-src/lib/`
- OpenSpec requirements under `ci-schema-coverage-rotation-toolchain` (canonical spec updated in place; no new capability id)
