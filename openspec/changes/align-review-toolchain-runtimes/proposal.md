## Why

The review workflow currently describes a mixed runtime bootstrap model that still includes `actions/setup-go` and a Node setup driven by the package engine range, even though the workflow source already carries explicit runtime declarations. This makes the requirement harder to reason about and leaves the version-alignment checks split between the workflow and repository declarations.

## What Changes

- Update the review-environment requirement for `ci-aw-openspec-verification` to require only explicit workflow runtimes of `go 1.26.1` and `node 24`.
- Remove the requirement to provision Go through `actions/setup-go` in the review environment.
- Require repository validation targets to check that the workflow's pinned Node runtime satisfies `package.json` `engines.node`, alongside the existing Go-to-`go.mod` alignment check.

## Capabilities

### New Capabilities
None.

### Modified Capabilities
- `ci-aw-openspec-verification`: Narrow the review workflow bootstrap requirement to explicit pinned Go and Node runtimes, drop `actions/setup-go`, and require make-target validation for both runtime declarations.

## Impact

Affected areas include the OpenSpec requirement for the verify-label workflow, the workflow source and compiled workflow artifacts, and the make-based checks that keep workflow runtime declarations aligned with `go.mod` and `package.json`.
