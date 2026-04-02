## Why

The review workflow change was headed toward explicit frontmatter runtime pins, but the desired maintenance model is the older repository-driven setup: Go should resolve directly from `go.mod`, Node should resolve directly from `package.json`, and the workflow should stop carrying a separate `runtimes.go` declaration. The agent still needs a deterministic handoff to the configured Go toolchain, so the workflow also needs to export `GOROOT` after Go setup for AWF chroot mode.

## What Changes

- Update the review-environment requirement for `ci-aw-openspec-verification` so Go is provisioned from `go.mod`, Node is provisioned from `package.json`, and `runtimes.go` is not used.
- Add a `Capture GOROOT for AWF chroot mode` step immediately after Go setup so the agent sees the configured Go toolchain.
- Remove the legacy Makefile runtime maintenance targets entirely now that the workflow reads repository version files directly.

## Capabilities

### New Capabilities
None.

### Modified Capabilities
- `ci-aw-openspec-verification`: Return the review workflow bootstrap to repository-driven Go and Node setup, add `GOROOT` export after Go installation, and stop using `runtimes.go`.
- `makefile-workflows`: Remove the legacy verify-label runtime maintenance targets and their supporting requirement text.

## Impact

Affected areas include the OpenSpec requirements for the verify-label workflow and lint/check behavior, the workflow source and compiled workflow artifacts, and the Makefile cleanup needed to drop the legacy runtime maintenance path.
