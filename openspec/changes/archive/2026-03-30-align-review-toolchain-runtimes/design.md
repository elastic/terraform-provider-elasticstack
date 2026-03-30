## Context

The verify-label workflow change currently assumes explicit runtime declarations in workflow frontmatter. The requested direction is to restore the repository-driven bootstrap model instead: `actions/setup-go` should read `go.mod`, `actions/setup-node` should read `package.json`, and the workflow should stop carrying a separate `runtimes.go` declaration. Because the AWF agent can execute in chroot mode, the workflow also needs to export `GOROOT` after Go setup so the configured Go toolchain remains visible to agent-executed commands.

This change spans the OpenSpec requirements, the workflow source and compiled output, and the Makefile cleanup needed to remove the old explicit-runtime maintenance path. Because the request changes both the workflow bootstrap contract and the surrounding maintenance story, a short design is useful before implementation.

## Goals / Non-Goals

**Goals:**
- Define the review workflow requirement around repository-driven runtime setup only.
- Provision Go from `go.mod` and Node from `package.json`.
- Remove `runtimes.go` from the authored workflow.
- Export `GOROOT` immediately after Go setup for AWF chroot mode.
- Preserve Terraform CLI availability requirements for the review environment.

**Non-Goals:**
- Changing the repository's declared Go version in `go.mod`.
- Changing the repository's declared Node engine range in `package.json`.
- Redesigning unrelated workflow verification, review, archive, or label-cleanup behavior.

## Decisions

### Use repository version files as the only runtime source of truth

The modified requirement will treat `go.mod` and `package.json` as the only maintainer-managed runtime declarations for the review environment. The workflow source should configure `actions/setup-go` with `go-version-file: go.mod` and `actions/setup-node` with `node-version-file: package.json`, instead of duplicating those versions in workflow frontmatter.

Alternative considered: keep explicit `runtimes.go` / `runtimes.node` pins and validate them against repository files. Rejected because it duplicates intent, creates sync work, and is unnecessary when the setup actions can already read the repository files directly.

### Export GOROOT after Go setup

The workflow should capture `go env GOROOT` into `GITHUB_ENV` immediately after `actions/setup-go` runs. This preserves access to the configured Go toolchain for AWF chroot mode without reintroducing a frontmatter Go runtime pin.

Alternative considered: rely on `actions/setup-go` alone with no extra environment export. Rejected because the compiled workflow already demonstrates that chroot mode benefits from an explicit `GOROOT` handoff, and the user requested that behavior be preserved.

### Remove the legacy runtime maintenance path entirely

The dedicated verify-label runtime maintenance targets were introduced only to maintain `runtimes.go.version` in workflow frontmatter. Once the workflow reads `go.mod` and `package.json` directly, that maintenance path should be removed from the Makefile, from `check-lint`, and from the related requirements text rather than retained as an explicit "not required" behavior.

Alternative considered: keep the targets around as extra belt-and-suspenders validation. Rejected because they would no longer validate the real runtime source of truth and would instead preserve maintenance complexity from the abandoned pinned-runtime approach.

## Risks / Trade-offs

- Relying on `node-version-file: package.json` follows the setup action's documented precedence across package metadata rather than a hardcoded major -> Mitigation: document that the repository file is the source of truth and keep workflow tests/generation checks in the normal validation path.
- Removing `runtimes.go` changes how maintainers reason about the agent environment -> Mitigation: update the requirement text and workflow comments together so the repository-driven model and `GOROOT` export are explicit.
- Removing the legacy runtime maintenance path could surprise contributors who learned the previous flow -> Mitigation: remove the stale comments, Makefile targets, and supporting requirement text in the same change so there is only one supported flow.
