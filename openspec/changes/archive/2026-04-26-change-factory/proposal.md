## Why

Maintainers currently have an agentic path for issue-driven implementation, but not for issue-driven OpenSpec proposal creation. A focused `change-factory` workflow would let trusted labeled issues become reviewable OpenSpec change pull requests without provisioning the Elastic Stack or asking the agent to implement code.

## What Changes

- Add a GitHub Agentic Workflow that reacts to trusted issue events labeled `change-factory`.
- Reuse the deterministic intake shape from `code-factory`: label qualification, actor trust checks, duplicate linked pull request suppression, deterministic branch naming, and a single linked pull request contract.
- Instruct the agent to create an OpenSpec change proposal from the issue title/body, including `proposal.md`, `design.md`, `tasks.md`, and required delta specs under `openspec/changes/<id>/`.
- Keep the workflow proposal-focused: it must not implement provider behavior, start the Elastic Stack, or run acceptance tests.
- Bootstrap only the toolchains needed to author and validate OpenSpec artifacts, with Node/OpenSpec as required and Go only if repository workflow-generation or validation commands require it.
- Defer any interactive exploration via GitHub comments or Discussions to a future workflow/state rather than including it in the initial `change-factory` behavior.

## Capabilities

### New Capabilities

- `ci-change-factory-issue-intake`: Defines the issue-labeled agentic workflow that creates exactly one linked OpenSpec change proposal pull request from a trusted GitHub issue.

### Modified Capabilities

- None.

## Impact

- New workflow source under `.github/workflows-src/` and generated workflow artifacts under `.github/workflows/`.
- New deterministic workflow helper logic and tests for trigger qualification, trust/duplicate gating, and linked pull request identification, likely adapted from the existing `code-factory` helpers.
- Agent prompt contract for creating and validating OpenSpec change artifacts from issue context.
- CI/workflow generation paths may need updates so the new workflow source is compiled and checked with existing workflow tooling.
