## 1. Update the Purpose line in the canonical spec (not delta-representable)

- [x] 1.1 In `openspec/specs/ci-build-lint-test/spec.md`'s `## Purpose` section, remove the phrase
      "preflight gate" from the sentence describing what the workflow defines. OpenSpec delta specs
      cannot express a `## Purpose` change for an existing capability (the delta's own `## Purpose` is
      ignored at sync time), so this edit must be made directly on the canonical spec file when this
      change is applied/synced, alongside applying the delta below.

## 2. Apply the delta spec to the canonical `ci-build-lint-test` capability

- [x] 2.1 Apply the `## MODIFIED Requirements` block for "Workflow identity and triggers (REQ-001–REQ-006)": update the workflow name/triggers to match `.github/workflows/provider.yml` (name `Provider CI`, `push` on `main`, `pull_request` types `[opened, synchronize, reopened]`, no `ready_for_review`/path-based exclusions), and reword the push scenario to reference the change-classification job instead of a preflight gate.
- [x] 2.2 Apply the `## REMOVED Requirements` entry for "Preflight gate (REQ-023–REQ-027)": delete the
      requirement and its three scenarios.
- [x] 2.3 Apply the `## REMOVED Requirements` entry for "Ready-for-review behavior (REQ-030)": delete
      the requirement and its scenario.
- [x] 2.4 Apply the `## MODIFIED Requirements` block for "Job permissions (REQ-028–REQ-029)": rename
      "preflight gate job" to "change-classification job" in the requirement text, and rename the
      scenario title from "Preflight permissions" to "Change-classification permissions" (its
      `GIVEN` clause names the change-classification job).
- [x] 2.5 Apply the `## MODIFIED Requirements` block for "Change classification gate
      (REQ-032–REQ-033)": drop `should_run` conditioning, state the full non-impacting-path list
      (`CHANGELOG.md`, `openspec/`, `.agents/`, `.github/` except `provider.yml` itself), and add the
      new "Push event always classifies as provider-impacting" scenario.
- [x] 2.6 Apply the `## MODIFIED Requirements` block renaming "Test validation job (REQ-034–REQ-036)"
      to "Provider gate job (REQ-034–REQ-036)": rewrite the pass/fail rule to match `gateProvider()`
      (including `golangci-lint` as a fourth gate input) and add the new "Provider change with an
      unexpected skip" scenario.
- [x] 2.7 Apply the `## MODIFIED Requirements` block for "Auto-approve job (REQ-018–REQ-021)": depend
      on the renamed `gate` job instead of `Test Validation`, drop the `ready_for_review` carve-out
      (the existing "Auto-approve after satisfied validation" scenario title is kept, updated to
      reference the `gate` job result), and add the new "Auto-approve does not run when the gate
      fails" scenario.
- [x] 2.8 Apply the `## REMOVED Requirements` entry for "Generated changelog pull requests can reach
      auto-approve without full CI": delete the requirement and its two scenarios.
- [x] 2.9 Apply the `## REMOVED Requirements` entry for "Changelog-only bypass remains narrowly
      scoped": delete the requirement and its two scenarios.

## 3. Verify

- [x] 3.1 Run
      `OPENSPEC_TELEMETRY=0 ./node_modules/.bin/openspec validate resync-ci-build-lint-test-preflight-drift --type change`
      and resolve any reported issues.
- [x] 3.2 After sync, confirm `openspec/specs/ci-build-lint-test/spec.md` no longer contains the
      strings "preflight", "should_run", or "Test Validation", and that every remaining requirement
      matches an actual job, script, or condition in `.github/workflows/provider.yml` and its
      `.github/scripts/workflows/` scripts.
- [x] 3.3 Confirm `openspec/specs/ci-pr-auto-approve/spec.md` is unchanged by this change — the
      `generated-changelog` requirements there remain the sole owner of that bypass behavior.
