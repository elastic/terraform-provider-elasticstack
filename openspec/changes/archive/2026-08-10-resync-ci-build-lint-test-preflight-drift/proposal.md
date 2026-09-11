## Why

`openspec/specs/ci-build-lint-test/spec.md` still narrates a preflight gate job that emits a
`should_run` output and conditions `build`/`lint`/`test`, `Test Validation`, change-classification, and
a changelog-only CI bypass on that output. `.github/workflows/provider.yml` has no such job: there is
no `preflight` job and no `should_run` output anywhere in the workflow. The real terminal job is
`gate` ("Provider Gate"), which evaluates `classify`/`build`/`lint`/`golangci-lint`/`test` via
`gateProvider()` in `.github/scripts/workflows/lib/gate-provider.js`. The changelog-only CI bypass
described in the current spec is not implemented by `provider.yml` itself (there is no preflight/gate
skip path for `generated-changelog`). #4419 already aligned the "Acceptance test job structure" requirement with reality;
this change resolves the remaining preflight/`should_run` drift called out in issue #4420 so the
capability no longer describes a mechanism that does not exist.

## What Changes

- Purpose text: drop the "preflight gate" phrase from the capability overview.
- Canonical workflow-implementation pointer: update prose from `.github/workflows/test.yml` to
  `.github/workflows/provider.yml` (not delta-representable; edited on the canonical spec directly).
- Canonical `## Schema` YAML block: align triggers with `.github/workflows/provider.yml`
  (`push.branches: [main]`, drop `tags-ignore`/`paths-ignore` and prior `branches: ['**']`, and
  `pull_request.types` without `ready_for_review`) — also not delta-representable.
- Workflow identity scenario ("Push to feature branch"): reword away from "the preflight gate allows
  execution" to a phrasing that does not assume a gate job.
- Delete `### Requirement: Preflight gate (REQ-023–REQ-027)` and its three scenarios — no such job or
  `should_run` output exists in `provider.yml`.
- Delete `### Requirement: Ready-for-review behavior (REQ-030)` and its scenario — `provider.yml`'s
  `pull_request` trigger types are `[opened, synchronize, reopened]`; `ready_for_review` is not a
  trigger and `gateProvider()` has no event-type branching, so this requirement describes a skip path
  that cannot occur.
- Rename `### Requirement: Job permissions (REQ-028–REQ-029)`'s "preflight gate job" reference to the
  real job it describes, the change-classification job (`classify`); the stated permissions
  (`contents: read`, `pull-requests: read`) already match `classify`'s actual permissions block.
- Rewrite `### Requirement: Change classification gate (REQ-032–REQ-033)` to drop the `should_run`
  conditioning (the `classify` job runs unconditionally in `provider.yml`) and to accurately state the
  non-impacting-path rule implemented by `classifyChanges()`: `CHANGELOG.md`, any path under
  `openspec/`, any path under `.agents/`, and any path under `.github/` except
  `.github/workflows/provider.yml` itself — not just paths under `openspec/`. Also state that
  non-`pull_request` events (`push` and `workflow_dispatch`) skip file inspection entirely and
  hardcode `provider_changes=true`.
- Rename `### Requirement: Test validation job (REQ-034–REQ-036)` to a name matching the real terminal
  job, `Provider Gate` (job id `gate`), and rewrite its pass/fail rule to match `gateProvider()`:
  passes when `classify=false` and every one of `build`/`lint`/`golangci-lint`/`test` is `skipped`, or
  when all four succeed (regardless of classify result); fails on any `failure`/`cancelled` result, on
  an unexpected `skipped` result when `classify=true`, or on any other leftover combination including
  an unrecognised classify or job result value. This also surfaces `golangci-lint` as a gate input,
  which the current spec omits entirely.
- Rewrite `### Requirement: Auto-approve job (REQ-018–REQ-021)` to depend on the renamed gate
  requirement instead of `Test Validation`, and drop the `ready_for_review`/preflight-skip carve-out:
  `provider.yml`'s real `auto-approve` job condition is unconditionally
  `needs.gate.result == 'success'` for `pull_request` events, with no event-type branching.
- Delete `### Requirement: Generated changelog pull requests can reach auto-approve without full CI`
  and `### Requirement: Changelog-only bypass remains narrowly scoped` (and their scenarios) from this
  capability outright. The changelog-only CI bypass they describe does not exist in `provider.yml` as a
  preflight skip path; the actual `generated-changelog` bypass is owned end-to-end by the
  `ci-pr-auto-approve` capability (`openspec/specs/ci-pr-auto-approve/spec.md`, "Generated changelog
  selector"/"commit authors"/"file allowlist" requirements), which already documents it correctly.

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `ci-build-lint-test`: remove every requirement, scenario, and prose reference that describes a
  preflight gate, a `should_run` output, a `ready_for_review` preflight-skip path, or a
  preflight-conditioned changelog-only CI bypass, none of which exist in `.github/workflows/provider.yml`.
  Rename and correct the requirements that describe the real terminal gate job (`gate`,
  "Provider Gate") and the real change-classification job (`classify`) so the capability matches the
  workflow's actual jobs and conditions.

## Impact

- `openspec/specs/ci-build-lint-test/spec.md` only — this is a documentation-only OpenSpec change.
- No changes to `.github/workflows/provider.yml`, its scripts, or any other workflow file.
- No changes to `openspec/specs/ci-pr-auto-approve/spec.md`; the `generated-changelog` bypass
  requirements already there are the sole owner of that behavior going forward.
- No Terraform provider Go code, generated clients, or docs outside `openspec/` are touched.
