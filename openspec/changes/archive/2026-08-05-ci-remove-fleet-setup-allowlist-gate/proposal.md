## Why

The Matrix Acceptance Test job's `setup-fleet` step in `.github/workflows/provider.yml` gates
`make setup-kibana-fleet` on an explicit per-version `if:` allowlist. `9.4.2` is a plain (non-snapshot)
matrix entry missing from that list even though every other nearby GA entry is present, so
Fleet-dependent acceptance tests silently skip Fleet bootstrap for `9.4.2` with no test-failure signal.

Investigation (see issue #4415) shows adding `9.4.2` to the list is the wrong fix: the step is
redundant everywhere it currently runs. CI already starts the stack with `make docker-fleet`, whose
`fleet_settings` service creates the default Fleet Server host, agent policy, and `fleet_server`
package policy before the agent starts. CI logs on versions where `setup-fleet` still runs show the
agent-policy and package-policy calls return `409 already exists`, and download-source bootstrapping
is handled separately by `acctest.PreCheck` via `ensureFleetDefaultAgentDownloadSource`. The allowlist
does not track a real stack capability boundary — it was introduced in #797 (Oct 2024) when CI used
GHA service containers with no `fleet_settings` equivalent, and has been maintained by habit (bump pin
/ promote SNAPSHOT) rather than by necessity since CI switched to `docker-compose` / `make docker-fleet`
in #1840 (Mar 2026). `8.18` and `8.19` were never added to the list either, and acceptance tests pass
green on all of them today.

Per the issue's recommendation, the fix is to remove the redundant step from the Matrix Acceptance
Test job rather than extend the allowlist, so future GA promotions (including the pending 9.5.0
promotion in #4403/#4404) stop needing to remember to update this list at all.

## What Changes

- Remove the `setup-fleet` step (`id: setup-fleet`, `name: Setup Fleet`, `run: make setup-kibana-fleet`)
  from the Matrix Acceptance Test job in `.github/workflows/provider.yml`. The job already provisions
  Fleet Server host, agent policy, and package policy via `make docker-fleet` (`fleet_settings`
  service), and default-agent-download-source bootstrap is already covered by `acctest.PreCheck`.
- Do **not** add `9.4.2` (or any other version) to the removed step's former allowlist — the fix is
  deletion of the gate, not extension of it.
- Leave `make setup-kibana-fleet` itself, and its other current callers
  (`.github/workflows/copilot-setup-steps.yml`, `.github/workflows/shared/elastic-stack.md`), unchanged.
  This change only removes the Matrix Acceptance Test job's redundant invocation; it does not audit or
  change those other invocations.
- Update the `ci-build-lint-test` capability to stop describing Fleet setup as running "only for
  configured version subsets" for the acceptance test job, since the job no longer has a per-version
  Fleet setup step.

## Capabilities

### Modified Capabilities
- `ci-build-lint-test`: acceptance test job no longer runs a per-version-gated Fleet setup step; Fleet
  bootstrap for the matrix job comes entirely from `make docker-fleet` plus `acctest.PreCheck`.

## Impact

- `.github/workflows/provider.yml`: `test` job (Matrix Acceptance Test) — one step removed.
- No provider Go code, generated clients, Makefile targets, or documentation are changed.
- No change to `.github/workflows/copilot-setup-steps.yml` or `.github/workflows/shared/elastic-stack.md`
  — both invoke `make setup-kibana-fleet` unconditionally outside the matrix job and are out of scope.
- CI acceptance coverage is unaffected: the step being removed was already redundant (confirmed by
  `409 already exists` responses in CI logs) everywhere it currently runs, and was already skipped for
  `8.18`, `8.19`, and `9.4.2` with acceptance tests passing.
