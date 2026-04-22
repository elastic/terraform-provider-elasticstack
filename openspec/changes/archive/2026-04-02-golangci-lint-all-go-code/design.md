## Context

The current canonical `makefile-workflows` spec says the `golangci-lint` target only lints code under `internal/`, but the root `Makefile` already invokes golangci-lint against `./...`. That mismatch leaves the published requirement narrower than the implemented contributor and CI behavior.

This change is intentionally small and localized. It updates the `makefile-workflows` contract so the documented lint scope matches the repository's actual Go module layout and the Makefile behavior used by `make lint` and `make check-lint`.

## Goals / Non-Goals

**Goals:**
- Define `golangci-lint` as repository-wide linting over `./...`.
- Preserve the existing distinction between `lint` fix mode and `check-lint` check-only mode.
- Keep the spec aligned with the existing Makefile behavior so contributors and CI share the same contract.

**Non-Goals:**
- Changing which linters are enabled in `.golangci.yaml`.
- Redesigning `lint` / `check-lint` ordering beyond the scope statement they already inherit.
- Expanding requirements for non-Go validation steps such as docs, formatting, workflows, or OpenSpec checks.

## Decisions

### Treat `./...` as the observable lint scope

The requirement should describe the `golangci-lint` target in terms of the Go module-wide package pattern `./...`, because that is the observable invocation contract contributors and CI rely on. This makes top-level packages such as `provider/`, `scripts/`, `xpprovider/`, and the module root part of the lint surface instead of implying that only `internal/` matters.

Alternative considered: describe the scope generically as "all repository Go code" without naming `./...`. Rejected because the user explicitly requested the `./...` contract, and the concrete package pattern is the clearest externally visible behavior.

### Preserve config-driven exclusions rather than restating them in the spec

The spec should require repository-wide linting while allowing golangci-lint's configured exclusions to continue shaping which paths are actually analyzed. That keeps the contract aligned with `.golangci.yaml` without hardcoding every excluded path into the requirement text.

Alternative considered: enumerate excluded directories in the requirement. Rejected because exclusions are configuration details that may evolve independently, while the stable contract is that the target runs from `./...` under repository configuration.

## Risks / Trade-offs

- Repository-wide linting may surface issues in packages outside `internal/` that were not previously captured by the old spec text -> Mitigation: this is the intended correction, and the Makefile already reflects the broader scope today.
- Referring to `./...` depends on Go package discovery semantics rather than a hand-maintained directory list -> Mitigation: that is the standard Go module-wide contract and better matches contributor expectations than a narrower hardcoded path.
