## Context

The `ci-aw-openspec-verification` workflow currently prepares Go with an explicit `actions/setup-go` step that reads `go.mod` via `go-version-file`. That step configures the runner environment, which is the context used by deterministic setup steps such as dependency installation and `make setup`. The `gh-aw` agent environment has a separate runtime model. The published `gh-aw` frontmatter reference documents `runtimes.go.version` as a required version string and does not document a `go-version-file` equivalent, expression-based indirection, or a way to import the Go version from repository files.

This change is limited to the workflow contract for review-environment setup plus the repository maintenance targets needed to keep the workflow frontmatter aligned with `go.mod`. It does not alter the workflow's label gating, review semantics, archive behavior, cleanup flow, or Renovate post-upgrade behavior.

## Goals / Non-Goals

**Goals:**
- Add `runtimes.go.version` for the agent environment without removing the existing runner-level `actions/setup-go` bootstrap.
- Capture the current documented limitation that the workflow must carry an explicit Go version value rather than reading `go.mod` directly through `runtimes:`.
- Preserve the existing Node, Terraform, `actions/setup-go`, and repository bootstrap expectations around `make setup`.
- Make future maintenance expectations explicit so Go version bumps update both `go.mod` and the workflow frontmatter together.
- Add repository-maintained check and sync targets so contributors have both a validation path and a repair path for Go runtime drift.

**Non-Goals:**
- Changing Node runtime selection, Terraform setup, or other frontmatter unrelated to Go provisioning.
- Wiring the sync target into `make renovate-post-upgrade` or changing Renovate configuration in this change.
- Broadly redesigning `ci-aw-openspec-verification` beyond the Go runtime source.

## Decisions

Use both Go runtime mechanisms because they serve different environments.
The workflow should keep `actions/setup-go` with `go-version-file: go.mod` for the runner workspace, because deterministic setup steps and dependency installation run there. The workflow should also declare `runtimes.go.version` in frontmatter so the agent environment uses the intended Go version during review work.

Alternative considered: use only `actions/setup-go`.
Rejected because it leaves the agent environment unspecified when review work runs inside the `gh-aw` runtime model.

Alternative considered: use only `runtimes.go.version`.
Rejected because it would stop configuring Go in the runner workspace where repository setup and dependency installation occur.

Treat the frontmatter runtime as an explicit value that must stay synchronized with `go.mod`.
Because the published `gh-aw` docs only document a literal `version` field for Go runtimes, the design should assume maintainers need to update that value deliberately when the repository Go version changes. The spec should require sync with `go.mod` so the agent environment does not silently drift from the runner environment and repository toolchain.

Alternative considered: assume `runtimes.go.version` can reference `go.mod` indirectly.
Rejected because the frontmatter reference documents only a version string and provides no file-based or external-version mechanism to rely on.

Add separate Makefile targets for check and sync.
The repository should expose one target that validates `go.mod` and `runtimes.go.version` match, and another target that rewrites the workflow frontmatter from `go.mod` and regenerates the compiled workflow. `check-lint` should call the validation target so CI and local verification catch drift automatically, while the sync target remains an explicit maintenance action.

Alternative considered: rely only on a `check-lint` failure.
Rejected because that detects drift but does not provide a standard repair path for contributors.

Alternative considered: rely only on a sync target.
Rejected because drift could still go unnoticed in CI or local review if contributors forget to run it.

Keep the rest of the bootstrap contract unchanged.
Node should continue following `package.json` engines, Terraform should remain available without wrapper behavior, `actions/setup-go` should continue preparing the runner environment from `go.mod`, and `make setup` should still prepare the agent workspace. This keeps the change focused on adding the missing agent runtime declaration instead of reopening the broader review-environment design.

Alternative considered: redesign all runtime provisioning around explicit frontmatter versions.
Rejected because the user request is specific to Go and the current Node/Terraform behavior is not under review here.

## Risks / Trade-offs

- Manual sync burden between workflow frontmatter and `go.mod` -> Add an explicit requirement that both stay aligned, and update tasks to check the frontmatter value against `go.mod`.
- New target logic may require fragile parsing of `go.mod` or workflow frontmatter -> Keep the sync/check implementation narrow and targeted to the known `go` line and `runtimes.go.version` key in the workflow source.
- Future `gh-aw` runtime features may later support file-based Go version resolution -> Keep the design narrowly framed around the currently documented behavior so it can be relaxed in a later change.
- Hardcoding an exact agent version may require more frequent workflow edits than `go-version-file` alone -> Accept the maintenance overhead in exchange for using the supported frontmatter model while preserving runner bootstrap from `go.mod`.

## Migration Plan

- Update the `ci-aw-openspec-verification` delta spec so Go provisioning is defined in terms of both runner setup through `actions/setup-go` and agent runtime configuration through explicit `runtimes.go.version` kept in sync with `go.mod`.
- Update the `makefile-workflows` delta spec to require drift-check and sync targets, and to require `check-lint` to invoke the drift check.
- Implement the new Makefile targets without changing `make renovate-post-upgrade`.
- Update `.github/workflows/openspec-verify-label.md` to add `runtimes.go.version` while preserving the explicit `actions/setup-go` step.
- Recompile `.github/workflows/openspec-verify-label.lock.yml` with `gh aw compile`.
- Verify the resulting workflow still prepares the review environment successfully and reflects the intended Go version.

## Open Questions

- None. Based on the current `gh-aw` frontmatter documentation, the agent runtime still appears to require a hardcoded Go version today even though the runner can continue reading `go.mod` through `actions/setup-go`.
