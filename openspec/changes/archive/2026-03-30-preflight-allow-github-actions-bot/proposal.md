## Why

The preflight gate currently recognizes only the Copilot coding agent email when deciding whether push-triggered CI may still run for branches that already have an open pull request. GitHub Actions-created commits use a different noreply address, so automation-originated pushes can be skipped even when they should be treated like bot-authored maintenance updates.

## What Changes

- Update the `ci-build-lint-test` preflight gate requirements so the "all commits are bot-authored" allowance accepts both the Copilot coding agent email and the GitHub Actions bot email `41898282+github-actions[bot]@users.noreply.github.com`.
- Keep the existing open-pull-request check, non-`push` event behavior, and `should_run` gating for downstream jobs unchanged.
- Clarify the affected scenarios so the allowed-author list is explicit anywhere the current Copilot-only rule is described.

## Capabilities

### New Capabilities

- _(none)_

### Modified Capabilities

- `ci-build-lint-test`: Expand the preflight gate's allowed bot author list for push events with open pull requests.

## Impact

- **Specs**: `openspec/specs/ci-build-lint-test/spec.md`
- **Workflow logic**: `.github/workflows/test.yml` preflight author validation must align with the updated allowed-user list.
- **CI behavior**: Bot-authored pushes from either Copilot or GitHub Actions continue to run push CI even when an open pull request already exists for the branch.
