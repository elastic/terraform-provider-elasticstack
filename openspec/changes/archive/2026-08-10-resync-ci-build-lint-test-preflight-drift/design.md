## Context

See `proposal.md` - Why for the drift this change resolves. The relevant real-world state, confirmed
by reading the workflow and its scripts directly:

- `.github/workflows/provider.yml` jobs: `classify` (Change Classification), `build`, `lint`,
  `golangci-lint` (Go Lint), `test` (Matrix Acceptance Test), `gate` (Provider Gate), `auto-approve`.
  There is no `preflight` job and no `should_run` output anywhere in the file.
- `classify` runs unconditionally (no `if:`) and always outputs `provider_changes`. Its permissions are
  `contents: read`, `pull-requests: read`.
- `build`, `lint`, `golangci-lint`, `test` each gate on
  `if: needs.classify.outputs.provider_changes == 'true'`.
- `gate` (`needs: [classify, build, golangci-lint, lint, test]`, `if: always()`) runs
  `.github/scripts/workflows/lib/runners/gate.js`, which dispatches to `gateProvider()` in
  `.github/scripts/workflows/lib/gate-provider.js`. `gateProvider()` passes when either
  `classify=false` and `build`/`lint`/`golangci-lint`/`test` are all `skipped`, or `classify=true` and
  all four are `success`; it fails on any `failure`/`cancelled` result, or on an unexpected `skipped`
  result when `classify=true`.
- `auto-approve` (`needs: [gate]`) runs when
  `always() && github.event_name == 'pull_request' && needs.gate.result == 'success'` - no
  `ready_for_review` or preflight-skip branching.
- `classifyChanges()` (`.github/scripts/workflows/lib/classify-changes.js`) treats a changed file as
  non-impacting when it is `CHANGELOG.md`, or starts with `openspec/`, or starts with `.agents/`, or
  starts with `.github/` and is not exactly `.github/workflows/provider.yml`. `provider_changes=true`
  when any changed file falls outside that set, or the file list is empty. For `push` events, the
  classifier script skips file inspection entirely and unconditionally sets `provider_changes=true`
  (`.github/scripts/workflows/provider/classify-changes.js`).
- The `generated-changelog` CI-bypass behavior described in the current `ci-build-lint-test/spec.md`
  (branch name, author, file-allowlist gates for reaching auto-approve without full CI) is not
  implemented by `provider.yml` itself: nothing in the workflow special-cases this combination as a
  preflight or gate skip path. If a generated-changelog bypass is intended, it should be owned (and
  implemented) under the `ci-pr-auto-approve` capability rather than duplicated here.

`@tobio` resolved both prior blocking questions on this issue in comments: delete the two
changelog-bypass requirements from `ci-build-lint-test` outright (ownership stays fully in
`ci-pr-auto-approve`), and rename the requirement that currently reads `Test Validation` to match the
real terminal job rather than keep a name-agnostic placeholder.

## Goals / Non-Goals

**Goals:**
- Leave no requirement, scenario, or prose line in `ci-build-lint-test/spec.md` that describes the
  preflight job, the `should_run` output, a `ready_for_review` preflight-skip path, or a
  preflight-conditioned changelog bypass as if it exists in `provider.yml` today.
- Rename and correct the requirements that cover the real `classify` and `gate` jobs so their names,
  dependencies, and pass/fail rules match `classify-changes.js` and `gate-provider.js`.
- Keep every requirement this issue does not name untouched (build/lint contents, matrix versions,
  pre-pull retry, snapshot notice, teardown, supply-chain pinning, and the already-aligned acceptance
  test job structure requirement from #4419).

**Non-Goals:**
- Any change to `.github/workflows/provider.yml` or its scripts - this is a docs-only OpenSpec change.
- Any change to `openspec/specs/ci-pr-auto-approve/spec.md`. Its `generated-changelog` requirements
  already correctly describe the bypass; this change only stops `ci-build-lint-test` from also
  claiming ownership of it via a nonexistent preflight skip path.
- A standalone `golangci-lint` job-behavior requirement (setup, version, args). It is only referenced
  as one of the four inputs to the rewritten gate requirement; documenting that job end-to-end remains
  a candidate follow-up.

## Decisions

### 1. Delete rather than reword the preflight-gate and ready-for-review requirements

`### Requirement: Preflight gate (REQ-023-REQ-027)` and `### Requirement: Ready-for-review behavior
(REQ-030)` are removed outright, not reworded into something else.

Why: both describe a job/output/skip-path combination that has no counterpart anywhere in
`provider.yml`. `provider.yml`'s `pull_request` trigger types are `[opened, synchronize, reopened]` -
`ready_for_review` is not even a configured trigger, and `gateProvider()` has no event-type branching.
There is no adjacent real behavior to reword these into; the correct spec state is that they do not
exist.

Alternatives considered: keep a trimmed-down version describing only the parts that might map to
`classify`'s unconditional run - rejected, because `classify` already has its own accurately-scoped
requirement, and keeping any preflight-shaped language invites the same drift to recur.

### 2. Rename, don't just reword, the terminal-gate and permissions requirements

`### Requirement: Test validation job (REQ-034-REQ-036)` is renamed to describe the `gate` job
("Provider Gate") directly, and `### Requirement: Job permissions (REQ-028-REQ-029)`'s "preflight gate
job" reference becomes "change-classification job" (`classify`).

Why: `@tobio` explicitly resolved this ("Rename it") rather than asking for a name-agnostic
description. The renamed gate requirement's pass/fail rule is rewritten from the current
disjunction-of-three-cases prose to mirror `gateProvider()`'s actual branches (classify-false-all-
skipped, classify-true-all-success, any-failure-or-cancelled, unexpected-skip-when-classify-true), and
gains `golangci-lint` as a fourth gate input, which the current spec omits.

Alternatives considered: keep the `Test Validation` name as a stable, implementation-agnostic label for
whatever job is terminal - rejected per `@tobio`'s explicit direction, and because a name detached from
the actual job (`gate`, "Provider Gate") is part of what let this drift accumulate undetected.

### 3. Delete the changelog-bypass requirements outright; do not move or merge their text

`### Requirement: Generated changelog pull requests can reach auto-approve without full CI` and
`### Requirement: Changelog-only bypass remains narrowly scoped` are deleted from `ci-build-lint-test`
with no replacement text in this capability.

Why: `@tobio` explicitly resolved this ("Delete them"). The behavior they describe is already fully
and correctly owned by `openspec/specs/ci-pr-auto-approve/spec.md`'s "Generated changelog selector" /
"commit authors" / "file allowlist" requirements, which gate the `generated-changelog` auto-approve
category - not a `ci-build-lint-test`-level CI skip path. Duplicating that ownership here (even in
corrected form) would reintroduce a second source of truth for the same behavior.

Alternatives considered: rewrite them to describe the bypass as auto-approve category routing instead
of a preflight skip - rejected per `@tobio`'s explicit "delete" direction, and because
`ci-pr-auto-approve` already documents this correctly, making a rewritten duplicate here pure
redundancy.

### 4. Rewrite the change-classification requirement's scope list and push-event behavior

`### Requirement: Change classification gate (REQ-032-REQ-033)` is rewritten to drop `should_run`
conditioning, state the full non-impacting-path list (`CHANGELOG.md`, `openspec/`, `.agents/`, and
`.github/` except `provider.yml` itself - not just `openspec/`), and state that `push` events hardcode
`provider_changes=true` without inspecting files.

Why: this is adjacent drift found while reading `classify-changes.js` directly, not just the five
locations the issue names verbatim. Leaving the scope list under-described would leave the spec
contradicting the same script the rewritten requirement is otherwise being corrected to describe
accurately.

### 5. Requirement ID convention: leave gaps, do not renumber

Deleted requirements' `REQ-NNN` ranges are left as gaps rather than renumbering later requirements to
close them.

Why: `openspec/specs/ci-pr-auto-approve/spec.md` already mixes numbered requirements (`REQ-001`
through `REQ-013`) with several newer requirements that carry no `REQ-NNN` suffix at all ("Renovate
selector", "Generated changelog selector", etc.) - the repo's convention tolerates non-contiguous and
even absent numbering rather than treating it as a stable, renumbered sequence. Renumbering here would
also churn every downstream cross-reference to the surviving `REQ-NNN` ids for no behavioral benefit.

## Risks / Trade-offs

- [A future reader searches for "preflight" in this capability and finds nothing, without a pointer to
  where the real gating logic now lives] -> Mitigation: the renamed gate and classification
  requirements name the actual jobs (`classify`, `gate`) and script files
  (`classify-changes.js`, `gate-provider.js`), so a reader who greps the workflow or scripts directory
  lands on the requirement that documents them.
- [Deleting the changelog-bypass requirements here could look like the behavior itself was removed]
  -> Mitigation: the proposal and this design explicitly state that `ci-pr-auto-approve` already owns
  and correctly documents that behavior; no functional change occurs.

## Migration Plan

This is a documentation-only change with no runtime migration. Sequence:

1. Edit `openspec/specs/ci-build-lint-test/spec.md` per the twelve edits enumerated in `tasks.md`
   (1.1–1.3 plus 2.1–2.9).
2. Run `OPENSPEC_TELEMETRY=0 ./node_modules/.bin/openspec validate resync-ci-build-lint-test-preflight-drift --type change`
   and resolve any reported issues.
3. Land the change; no code, workflow, or script file changes accompany it.

## Open questions

- Non-blocking: open a follow-up for a standalone `golangci-lint` job-behavior requirement (setup,
  version, args), or wait for a separate issue?
- Non-blocking: open a follow-up to resync `### Requirement: Build and lint jobs
  (REQ-007–REQ-008, REQ-031)` with the real `build`/`lint` jobs (and the separate `golangci-lint`
  job) in `.github/workflows/provider.yml` — left untouched by this change's Goals.
- Non-blocking: this repo's convention for `REQ-NNN` ID stability after deletions (renumber vs. leave
  gaps) — check `ci-pr-auto-approve` before drafting the diff.

Resolution recorded for the second question (non-blocking, but settled while drafting this change; see
Decision 5 above): `ci-pr-auto-approve` already tolerates non-contiguous and unnumbered requirements,
so this change leaves gaps rather than renumbering.

Note: the former open question about stale `test.yml` pointer / `Build/Lint/Test` name / push-branch
scope / Schema YAML was resolved during apply — those are aligned with `provider.yml` in the
canonical spec.
