## Context

The dead-code removal rotation workflow (`ci-deadcode-removal-rotation.md`) removes a single dead symbol per run, optionally removes companion tests, verifies with `make build` and `go test`, then opens a cleanup PR. Because the workflow does not run a formatter before committing, the resulting PR frequently triggers a CI lint failure on the `check-fmt` step, forcing manual intervention.

The repository's `make fmt` target runs `gofmt -w` and `goimports` in a single idempotent pass. Running it after file edits and before committing ensures the PR diff is already lint-clean on arrival.

## Goals / Non-Goals

**Goals:**
- Ensure every dead-code cleanup PR is free of formatting failures when it reaches CI.
- Keep the fix minimal: one additional step in the existing task sequence.

**Non-Goals:**
- Changing the pre-activation logic, selection algorithm, or memory-recording contract.
- Running `make fmt` on files outside the candidate and companion test files.
- Replacing the existing build and unit-test verification steps.

## Decisions

### 1. Run `make fmt` after verification, before PR creation

The formatting step is placed **after** the existing `make build` + `go test` verification and **before** the `create-pull-request` safe output. This ordering avoids running the formatter on code that would later be rolled back due to a build or test failure.

Why:
- Formatting is fast and idempotent; placing it last in the verification chain adds negligible overhead.
- Running it only after a successful build means we never format and then discard the result.

Alternatives considered:
- Run `make fmt` before `make build`: rejected because a formatter pass on broken code is wasted work and could obscure build error context.
- Run `make fmt` as part of pre-activation: rejected because pre-activation is deterministic and should not modify repository files.

## Risks / Trade-offs

- [`make fmt` targets may differ across Go/tool versions] — the workflow already pins Go via `actions/setup-go`; `make fmt` uses the same pinned toolchain, so drift is unlikely.
- [Extra step adds latency] — formatting is sub-second on small diffs; not a practical concern.

## Open Questions

None for the initial implementation.
