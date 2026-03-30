## Why

The `openspec-verify-label` workflow currently leaves key gating and setup work to the agent instructions: defensive label verification, pull request file inspection, active change selection, and OpenSpec CLI setup. Moving those actions into deterministic pre-agent steps makes activation reproducible, reduces agent ambiguity, and lets the agent start with a known selected change and ready-to-use tooling.

## What Changes

- Update the `ci-aw-openspec-verification` workflow contract so label verification, PR file inspection, active change selection, and OpenSpec CLI setup run as deterministic pre-activation steps instead of natural-language agent instructions.
- Require the workflow to compute the selected active change id before agent execution, skip the expensive agent job when deterministic gating fails, and expose the selected change and related gate results as pre-activation outputs for later jobs.
- Require repository-standard OpenSpec installation in the deterministic setup path so the agent can rely on `npx openspec` without re-performing setup logic.
- Narrow the agent instructions to verification, relevance review, review submission, and archive-on-approve behavior using the precomputed change context.

## Capabilities

### New Capabilities
<!-- None. -->

### Modified Capabilities
- `ci-aw-openspec-verification`: move workflow gating and OpenSpec setup into deterministic pre-agent steps, and allow the selected change id to be supplied to the agent from pre-activation outputs

## Impact

- `.github/workflows/openspec-verify-label.md`
- `.github/workflows/openspec-verify-label.lock.yml`
- Pre-activation workflow logic, including GitHub API access for PR files, step outputs consumed by later jobs, and job conditions that can skip agent execution
- OpenSpec verification runs that now depend on deterministic setup instead of agent-discovered setup
