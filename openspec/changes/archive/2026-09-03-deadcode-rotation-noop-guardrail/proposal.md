## Why

The Dead-code Removal Rotation workflow (`.github/workflows/ci-deadcode-removal-rotation.md`) produced no safe outputs on run [33585414784](https://github.com/elastic/terraform-provider-elasticstack/actions/runs/33585414784) (reported as issue #4767): the agent job completed without opening a pull request and without calling the `noop` safe output. The gh-aw framework flags this as a failure because a run with no safe output is indistinguishable from a broken agent.

The agent task's step 4 ("Verify") instructs the agent to record a failure reason and "stop without creating a PR" when `make build` fails/times out or when `go test` fails for the impacted packages — but it never explicitly instructs the agent to call `noop` on those paths. Only the companion-test backstop (step 3) and the catch-all closing sentence ("If you cannot safely proceed at any point, record the appropriate reason code and call `noop`") reference `noop` at all, and an agent that treats step 4's "stop" as the terminal instruction for that branch has no explicit reason to continue on to the catch-all. Other rotation-style workflows in this repository (e.g. `schema-coverage-rotation.md`) avoid this ambiguity by pairing every "no actionable outcome" branch with an explicit `MUST call noop` instruction inline, rather than relying on a single closing sentence to cover every branch.

## What Changes

- Make every "stop without opening a pull request" branch in the dead-code removal agent task (`make build` failure, verification timeout, `go test` failure) explicitly instruct the agent to call the `noop` safe output with a concise reason, immediately alongside the existing memory-recording instruction for that branch.
- Add an explicit Guardrail stating that the run MUST end with either a `create-pull-request` or a `noop` safe output call — the agent must never stop mid-task without emitting one of the two.

## Capabilities

### New Capabilities
<!-- None. -->

### Modified Capabilities
- `deadcode-removal-rotation`: require an explicit `noop` safe-output call on every verification-failure branch (build failure, verification timeout, test failure) that stops the run without opening a pull request, and add a run-level guardrail that every run must terminate with exactly one safe output (`create-pull-request` or `noop`).

## Impact

- `.github/workflows/ci-deadcode-removal-rotation.md` — update the agent task's verification-failure branches (step 4) to explicitly call `noop`, and add the run-level "must terminate with a safe output" guardrail. The compiled `.github/workflows/ci-deadcode-removal-rotation.lock.yml` must be regenerated after the source `.md` changes.
- No changes to pre-activation logic (`scripts/ci-deadcode-removal-rotation/*.go`), the memory schema, or the set of valid attempt-reason codes.
