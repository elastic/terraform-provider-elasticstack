## Context

`.github/workflows/provider.yml` (`test` job, "Matrix Acceptance Test") drives Elastic Stack version
coverage from a single `strategy.matrix.version` list. Snapshot handling is derived structurally from
the version string via `endsWith(matrix.version, '-SNAPSHOT')` rather than a separate flag:

- Fleet setup (`setup-fleet` step) runs when `matrix.version` is one of an explicit list of GA
  versions **or** `endsWith(matrix.version, '-SNAPSHOT')` is true.
- Acceptance tests (`tf-acceptance` step) use `continue-on-error: ${{ endsWith(matrix.version,
  '-SNAPSHOT') }}`, so only snapshot-labeled entries are allowed to fail without failing the job.
- A snapshot-only PR warning comment step fires on `tf-acceptance` failure when
  `endsWith(matrix.version, '-SNAPSHOT')`.

`9.5.0-SNAPSHOT` is currently the sole snapshot entry, tracking the in-development 9.5 line ahead of
its GA release. `9.5.0` has now shipped, so the matrix entry must become `9.5.0` and stop being
treated as a snapshot.

**Observed gap this design accounts for:** the Fleet setup step's explicit GA version list currently
reads `8.10.4, 8.11.4, 8.12.2, 8.13.4, 8.14.3, 8.15.5, 8.16.6, 8.17.10, 9.0.8, 9.1.10, 9.2.8, 9.3.6 ||
endsWith(matrix.version, '-SNAPSHOT')` — it does **not** include `9.4.2`, even though `9.4.2` is a
plain (non-snapshot) matrix entry newer than `9.3.6`. Git history shows `9.4.0` was added directly as
a GA entry (not promoted from a `9.4.x-SNAPSHOT` label), so it was never swept into this list, and the
later `9.4.0` → `9.4.2` patch bump didn't add it either. Whether or not that omission is intentional
for `9.4.2` specifically, it demonstrates the exact failure mode this change must avoid for `9.5.0`:
a version that relied on the `endsWith(..., '-SNAPSHOT')` fallback for Fleet coverage silently loses
that coverage the moment its matrix entry is rewritten to the GA string, unless it is added to the
explicit list in the same change.

## Goals / Non-Goals

**Goals:**
- Test the 9.5 line against the released `9.5.0` image instead of a snapshot build.
- Make `9.5.0` acceptance failures blocking, matching every other GA entry in the matrix.
- Preserve the Fleet setup coverage `9.5.0` had while labeled `9.5.0-SNAPSHOT`, by adding it to the
  explicit per-version condition list rather than relying on the snapshot suffix.
- Record a durable requirement so future snapshot-to-GA promotions (9.6, 9.7, ...) apply the same
  checklist instead of re-discovering the Fleet-coverage gap each time.

**Non-Goals:**
- Adding a new `9.6.0-SNAPSHOT` (or later) entry to start tracking the next in-development line. The
  issue only asks to promote the current 9.5 entry; starting the next snapshot line is a separate,
  independently-timed follow-up once that snapshot build exists and is worth testing.
- Auditing or fixing the pre-existing `9.4.2` Fleet-setup-list gap noted above. It predates this
  change and is not caused by it; calling it out here is scoping context for reviewers, not
  in-scope work. It may be worth a separate follow-up issue.
- Any change to `force-install-synthetics` gating — that step's explicit version list
  (`8.14.3, 8.15.5, 8.16.6, 8.17.10`) does not include any 9.x version today and is unaffected by this
  promotion.
- Provider Go code changes *unrelated to* making the newly-blocking `9.5.0` entry pass. This was
  scoped as a Non-Goal when the design was first written, before the promotion had actually run
  against real Kibana 9.5.0 GA. In practice, making `9.5.0` blocking surfaced three real acceptance
  regressions against the GA image (see Decision 4 below); fixing those *is* in scope, since they are
  the direct, causally-required consequence of this promotion — the `test` job cannot pass without
  them, and leaving them failing would defeat the purpose of removing `continue-on-error`. Provider
  code changes unrelated to those three regressions remain out of scope.

## Decisions

### 1. Rewrite the matrix entry in place rather than adding a parallel GA entry

Change `"9.5.0-SNAPSHOT"` to `"9.5.0"` at its existing position in `strategy.matrix.version`, matching
the exact pattern used for the prior `9.4.0-SNAPSHOT` → `9.4.0` promotion (see git history on
`.github/workflows/provider.yml`).

Why:
- Keeps one matrix entry per tracked stack line; avoids doubling CI runtime for the same version.
- Matches established repository precedent for this exact transition.

Alternatives considered:
- Keep both `9.5.0-SNAPSHOT` and add `9.5.0`: rejected — the snapshot label exists to track the next
  *unreleased* build; once 9.5.0 ships there is nothing left for the snapshot label to track until a
  9.6 snapshot exists.

### 2. Add the promoted version to the Fleet setup explicit list in the same change

Add `9.5.0` to the `setup-fleet` step's `if:` condition list of explicit versions, alongside the
matrix rewrite.

Why:
- Without this, `9.5.0` loses Fleet setup the moment it stops matching
  `endsWith(matrix.version, '-SNAPSHOT')`, silently narrowing acceptance coverage with no test
  failure signal (Fleet-dependent tests would simply not run rather than fail).
- This is the concrete mechanism behind the `9.4.2` gap described above; doing it explicitly here
  prevents 9.5.0 from landing in the same state.

Alternatives considered:
- Rely on a broader "9.x and above" pattern instead of an explicit per-version list: out of scope —
  that would be a larger refactor of the workflow's conditional structure, not a version bump, and
  risks changing behavior for versions not part of this issue.

### 3. Record the promotion checklist as an OpenSpec requirement

Add a `ci-build-lint-test` requirement describing the snapshot-to-GA promotion contract: rewrite the
matrix entry, drop the snapshot label, and re-add the version to any per-version step condition that
had matched it only via the snapshot suffix.

Why:
- The existing capability spec already describes snapshot vs. non-snapshot behavior abstractly
  ("snapshot versions allowed to fail... while non-snapshot versions remain blocking") but says
  nothing about what must happen at the moment of promotion, which is exactly where the Fleet-setup
  gap above went unnoticed.

Alternatives considered:
- Leave this as an unwritten convention: rejected — it already produced one silent coverage gap
  (`9.4.2`), and codifying the checklist is cheap.

### 4. Fix the three acceptance regressions this promotion surfaced, in this same change

Making `9.5.0` blocking (Decision 1) surfaced three real failures against the released Kibana 9.5.0
GA image, none of which are CI-matrix or workflow issues — all three are provider bugs or a genuine
Kibana-side API behavior change, confirmed by comparing CI history against `9.5.0-SNAPSHOT` and
against earlier GA versions:

1. **Null-preservation bug (10 panel types).** `PopulateFromAPI`'s "type-change recovery" branch
   checked `pm.<Type>Config == nil`, but callers (`dashboardMapPanelFromAPI` in `models_panels.go`)
   always pass a zero-valued `PanelModel` deliberately, to avoid aliasing plan pointers — so that
   check is always true and the branch fired on every same-type update, not just genuine type
   changes, skipping REQ-009 null-preservation entirely. Invisible while Kibana never returned a
   concrete value for an unset field; Kibana 9.5.0 GA started doing so for several enum-shaped
   fields (e.g. `aiops_pattern_analysis_config.minimum_time_range`), producing `Provider produced
   inconsistent result after apply`. Fixed by keying on `prior.<Type>Config` instead, which
   correctly distinguishes "this panel was already this type" (honor null intent) from "this panel
   just became this type" (no prior intent to honor) — `pm`'s own state can't make that
   distinction because it never carries it into `PopulateFromAPI` in the first place.
2. **`data_source_json` `name` key.** Kibana 9.5.0 GA started echoing a `name` key in
   `data_view_spec` payloads for `data_source_json` that earlier versions omitted, breaking
   apply-consistency for every Lens by-value chart type. Fixed by adding `"name"` alongside the
   existing `"time_field"` entry in the already-established
   `lenscommon.PreservePlanJSONIfStateAddsOptionalKeys` call sites — the same mechanism previously
   added for the analogous `time_field` case, extended rather than duplicated.
3. **`ml_anomaly_charts_config.severity_threshold` raw range.** Confirmed via CI run history that an
   arbitrary (non-canonical) `min`/`max` range passed against `9.5.0-SNAPSHOT` on `main` hours before
   this promotion landed, but the released `9.5.0` GA image rejects it with HTTP 400 unless the
   `{min, max}` pair exactly matches one of five fixed canonical pairs (the generated client models
   `severity_threshold` as a 5-member union pinning both `min` and `max` together, not `min` alone).
   This is a genuine Kibana-side API change
   between the pre-GA snapshot and the GA release, not a provider defect — the provider already
   faithfully passes through the configured range. Since `ml_anomaly_charts` requires Kibana
   `>=9.5.0-SNAPSHOT` (no earlier supported version), there is no fallback version this scenario
   could target instead. Removed the now-permanently-failing
   `TestAccResourceDashboardMlAnomalyChartsRawRange` test (kept
   `TestAccResourceDashboardMlAnomalyChartsRawRangeCanonicalCoincidence`, which covers a passing
   raw-range case) and updated the `kibana-dashboard` capability's REQ-053 to document the
   canonical-boundary constraint (see the `kibana-dashboard` delta spec in this change).

Why:
- All three are direct, causally-required consequences of Decision 1: the `test` job cannot pass for
  `9.5.0` without them, and re-adding `continue-on-error` to work around them would defeat the entire
  purpose of this promotion.
- Deferring them to a follow-up PR would mean landing this change with `9.5.0` still red, which is a
  worse outcome than fixing them now that they've been found and verified against a real 9.5.0 stack.

Alternatives considered:
- Land the CI-matrix change alone and file follow-up issues for the three regressions: rejected — the
  promotion's entire purpose is to make `9.5.0` failures blocking, so landing it while still red
  contradicts that purpose without a compelling reason to delay (the fixes were already found,
  understood, and verified before this decision was made).

## Risks / Trade-offs

- [Promoting to a fixed `9.5.0` string means a later 9.5.x patch is not automatically covered] ->
  Mitigation: this matches existing matrix conventions (every other GA entry is pinned to an exact
  patch), and patch bumps are handled the same way subsequent point releases already are.
- [Making `9.5.0` blocking could surface previously-masked acceptance failures] -> Mitigation: this is
  the intended effect of GA promotion (per the issue and per Decision 1); any failures surfaced are
  real coverage gaps that should be fixed or triaged, not suppressed by re-adding
  `continue-on-error`.

## Migration Plan

1. Update `.github/workflows/provider.yml`: change `"9.5.0-SNAPSHOT"` to `"9.5.0"` in
   `strategy.matrix.version`, and add `9.5.0` to the `setup-fleet` step's explicit version condition.
2. Confirm no other step or script keys off `endsWith(matrix.version, '-SNAPSHOT')` or an explicit
   version list in a way that also needs `9.5.0` added (per the audit in this design, only Fleet setup
   does).
3. Land the change; the next CI run against the `9.5.0` entry exercises the full blocking path.

## Open Questions (resolved)

- Should the pre-existing `9.4.2` Fleet-setup-list gap be fixed as part of this change, or tracked as
  a separate follow-up issue? **Resolved: tracked separately**, not fixed here. It predates and is
  independent of the 9.5 promotion (Non-Goals), and this change already has three in-scope regression
  fixes (Decision 4); bundling an unrelated pre-existing gap risks conflating review of the two.
  Filed as follow-up issue #4415.
- When should a `9.6.0-SNAPSHOT` (or later) entry be added to resume tracking the next in-development
  line? **Resolved: deferred**, per Non-Goals — a separate, independently-timed change once a `9.6`
  snapshot build exists and is worth testing. No action taken in this change.
