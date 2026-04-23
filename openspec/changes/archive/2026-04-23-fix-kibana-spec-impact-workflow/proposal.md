## Why

The `kibana-spec-impact` workflow currently splits deterministic pre-activation logic across an inline GitHub Script step, agent-only repo-memory initialization, and repo-root JSON files that do not survive the handoff into later jobs. That leaves the workflow brittle: pre-activation cannot reliably use repo memory, the report path is not durable between jobs, and the agent instructions point at files that are not guaranteed to exist in the agent environment.

## What Changes

- Move the pre-activation compute-and-gate flow behind the repository Go helper so the workflow does not depend on an inline `actions/github-script` step for Kibana spec-impact computation.
- Update `on.steps` to check out the repository before Go setup and to initialize repo memory during pre-activation rather than only in the agent job, using a dedicated checkout/init step against the same explicit repo-memory branch configured in the workflow.
- Upload the deterministic Kibana spec-impact report as a GitHub Actions artifact in pre-activation and download it into `/tmp/gh-aw/agent` for the agent job.
- Update the agent instructions to read the downloaded report artifact and to write any agent-produced JSON support files under `/tmp/gh-aw/agent`.
- Regenerate the derived workflow markdown and lockfile so the authored template and compiled workflow stay aligned.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `ci-kibana-spec-impact-issues`: tighten the workflow contract so pre-activation has access to repo memory through an explicit configured branch, deterministic impact evidence is handed off durably between jobs via artifact download into `/tmp/gh-aw/agent`, and the agent consumes the downloaded artifact paths rather than ephemeral repo-root files.

## Impact

- Workflow source under `.github/workflows-src/kibana-spec-impact/`
- Compiled workflow artifacts under `.github/workflows/`
- Kibana spec-impact helper CLI under `scripts/kibana-spec-impact/`
- Workflow generation and validation flows (`workflow-generate`, `make workflow-test`, `make check-workflows`)
