## Why

The `schema-coverage-rotation` workflow currently runs on the GitHub-hosted Copilot engine. The repository wants this workflow to follow the same LiteLLM routing direction as other GH AW migrations, but with the worker itself using the `claude` engine through an Anthropic-compatible proxy at `https://elastic.litellm-prod.ai/`.

## What Changes

- Update the `schema-coverage-rotation` workflow contract so the rotation worker uses `engine.id: claude`.
- Route Claude API traffic through the Elastic LiteLLM Anthropic-compatible endpoint via `ANTHROPIC_BASE_URL` and a secret-backed `ANTHROPIC_API_KEY`.
- Extend the workflow's AWF network policy to allow the LiteLLM host in addition to the existing bootstrap ecosystems.
- Require an explicit tool timeout suitable for the Claude engine so repository-local schema analysis commands are not constrained by Claude's shorter default tool-call budget.
- Preserve the existing issue-slot gating, repo-memory flow, issue creation rules, and downstream `assign-to-agent` behavior.

## Capabilities

### New Capabilities

- `ci-schema-coverage-rotation-engine`: Define the schema-coverage rotation workflow's Claude engine, Anthropic proxy configuration, and explicit Claude execution budget.

### Modified Capabilities

- `ci-schema-coverage-rotation-toolchain`: Extend the workflow network requirements so the existing bootstrap/toolchain contract also allows the LiteLLM proxy host needed by the Claude engine.

## Impact

- Authored workflow source under `.github/workflows-src/schema-coverage-rotation/`
- Generated workflow artifacts under `.github/workflows/schema-coverage-rotation.*`
- GH AW engine authentication and network policy for the schema-coverage worker
- Workflow generation tests under `.github/workflows-src/lib/`
- OpenSpec requirements for `ci-schema-coverage-rotation-engine` and `ci-schema-coverage-rotation-toolchain`
