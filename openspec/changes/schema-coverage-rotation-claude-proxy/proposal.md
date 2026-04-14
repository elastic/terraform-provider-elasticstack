## Why

The `schema-coverage-rotation` workflow currently runs on the GitHub-hosted Copilot engine. The repository wants this workflow to follow the same LiteLLM routing direction as other GH AW migrations, but with the worker itself using the `claude` engine through an Anthropic-compatible proxy at `https://elastic.litellm-prod.ai/`.

## What Changes

- Update the `schema-coverage-rotation` workflow contract so the rotation worker uses `engine.id: claude`.
- Route Claude API traffic through the Elastic LiteLLM Anthropic-compatible endpoint via `ANTHROPIC_BASE_URL` and a secret-backed `ANTHROPIC_API_KEY`.
- Extend the authored workflow `network.allowed` with `elastic.litellm-prod.ai` alongside the existing bootstrap ecosystems. The compiled Claude lock file still reflects a **broader** compiler-generated AWF domain bundle for agent execution (merged into AWF `--allow-domains`), not only the keys declared in YAML; this change accepts that reviewed trade-off.
- Require an explicit tool timeout suitable for the Claude engine so repository-local schema analysis commands are not constrained by Claude's shorter default tool-call budget.
- Preserve the existing issue-slot gating, repo-memory flow, issue creation rules, and downstream `assign-to-agent` behavior.

## Operator prerequisites

This workflow depends on the GitHub Actions repository secret **`CLAUDE_LITELLM_PROXY_API_KEY`**. Configure it in the repo (or organization) secrets used by Actions before relying on scheduled or manual runs.

## Capabilities

### New Capabilities

- `ci-schema-coverage-rotation-engine`: Define the schema-coverage rotation workflow's Claude engine, Anthropic proxy configuration, and explicit Claude execution budget.

### Modified Capabilities

- `ci-schema-coverage-rotation-toolchain`: Extend the workflow network requirements so the existing bootstrap/toolchain contract also allows the LiteLLM proxy host needed by the Claude engine.

## Impact

- Authored workflow source under `.github/workflows-src/schema-coverage-rotation/`
- Generated workflow artifacts under `.github/workflows/schema-coverage-rotation.*`
- GH AW engine authentication, new secret **`CLAUDE_LITELLM_PROXY_API_KEY`**, and effective AWF network surface (authored allowlist keys plus compiler-expanded domain set in the lock)
- Workflow generation tests under `.github/workflows-src/lib/`
- OpenSpec requirements for `ci-schema-coverage-rotation-engine` and `ci-schema-coverage-rotation-toolchain`
