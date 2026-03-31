## Why

The current `makefile-workflows` requirements say the `golangci-lint` target only scans `internal/`, which no longer matches the intended repository-wide linting contract for a Go module with code in multiple top-level packages. The spec should require `golangci-lint` to evaluate all Go packages under `./...` so contributor and CI validation cover the full repository.

## What Changes

- Update the `makefile-workflows` requirements so the `golangci-lint` target is specified against repository-wide Go code (`./...`) rather than only `internal/`.
- Clarify the observable behavior of `lint` and `check-lint` so both paths inherit that full-repository lint scope, with `lint` still enabling fix mode and `check-lint` remaining check-only.
- Align the Makefile implementation with the updated requirement so local and CI lint runs cover all Go packages in the module.

## Capabilities

### New Capabilities
None.

### Modified Capabilities
- `makefile-workflows`: Change the golangci-lint requirement from `internal/`-only linting to repository-wide `./...` linting for all Go code.

## Impact

Affected areas include the canonical `makefile-workflows` spec, the root `Makefile` lint recipes, and any contributor or CI workflows that rely on `make lint` or `make check-lint` for Go validation.
