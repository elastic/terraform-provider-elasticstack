## Context

The repository currently splits unit-style verification across multiple entry points. `make test` runs Go unit tests only, `make check-lint` also runs `workflow-test`, and `.agents/hooks/*.test.mjs` is not covered by a dedicated Makefile target. CI mirrors that split by running the build job without the workflow or hook tests, while the lint job carries some of the unit-style coverage.

This change is intentionally small, but it spans the root `Makefile`, JavaScript test locations, and the main CI workflow. A short design keeps the intended boundaries clear before implementation.

## Goals / Non-Goals

**Goals:**
- Make `make test` the single Makefile entry point for repository unit-style test execution.
- Add a named Makefile target for hook JavaScript tests so local and CI usage share the same interface.
- Keep the CI build job aligned with the Makefile by running the workflow and hook tests there.

**Non-Goals:**
- Changing acceptance test behavior or the `testacc` targets.
- Moving workflow generation freshness checks out of `check-lint`.
- Introducing new external test frameworks or non-Node tooling.

## Decisions

1. Add a dedicated `hook-test` Makefile target.
Rationale: the hook tests already exist as a coherent Node test suite under `.agents/hooks/`. Giving them a named target avoids embedding raw `node --test` commands into multiple aggregate targets and workflows, and keeps CI coupled to the Makefile contract instead of file globs.

Alternative considered: invoke `node --test .agents/hooks/*.test.mjs` directly from `make test` and CI. Rejected because it duplicates command details and makes future changes harder to centralize.

2. Expand `make test` into an aggregate target for all unit-style suites.
Rationale: contributors expect `make test` to represent the repository's unit-level verification. Making it depend on or invoke Go unit tests, `workflow-test`, and `hook-test` closes the current gap without changing acceptance-test behavior.

Alternative considered: leave `make test` focused on Go and create a second aggregate target. Rejected because it preserves the ambiguity the change is intended to remove.

3. Keep CI build-job execution aligned with the named Make targets.
Rationale: the build job should verify repository unit-style tests before or alongside compilation. Using `make workflow-test` and `make hook-test` in CI preserves a single source of truth for test commands and lets the Makefile remain the canonical developer contract.

Alternative considered: move these tests into the lint job only. Rejected because the requested behavior is to run them in CI build flow and to treat them as tests rather than lint checks.

## Risks / Trade-offs

- [Build job grows slightly longer] -> Mitigation: only lightweight Go/Node unit-style tests are added, not acceptance coverage.
- [Two places still mention workflow-related checks] -> Mitigation: keep `check-workflows` in lint for freshness validation while reserving `workflow-test` for executable tests.
- [Node runtime requirements expand in CI build] -> Mitigation: explicitly document Node setup in the build job requirements so implementation remains reproducible.
