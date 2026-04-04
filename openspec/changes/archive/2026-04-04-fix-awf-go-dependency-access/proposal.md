## Why

The `openspec-verify-label` workflow currently prepares Go dependencies during `make setup`, but agent-executed Go commands in AWF still attempt live module downloads during the sandboxed review phase. This causes avoidable verification failures because the workflow neither grants AWF Go ecosystem network access nor exports the Go cache paths needed for chroot-mode reuse of the prepared module cache.

## What Changes

- Allow the `openspec-verify-label` workflow's AWF network policy to use the Go ecosystem so agent-executed Go commands can resolve modules when cache reuse is insufficient.
- Export additional Go environment variables after `actions/setup-go` so AWF chroot mode can reuse the prepared Go workspace and module cache, not just the Go toolchain root.
- Clarify the review-environment requirement so prepared Go dependencies remain available to agent-invoked Go commands during verification.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `ci-aw-openspec-verification`: Change the review-environment requirements so the workflow grants AWF Go network access and exports the Go cache/workspace paths needed for chroot-mode dependency reuse.

## Impact

- Workflow source and compiled lock file for `.github/workflows/openspec-verify-label`
- Review-environment setup steps and AWF network policy
- OpenSpec requirements for `ci-aw-openspec-verification`
