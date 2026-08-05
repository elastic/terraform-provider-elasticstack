## Why

The `Provider CI` acceptance matrix in `.github/workflows/provider.yml` currently tracks the
in-development 9.5 line as `9.5.0-SNAPSHOT`. Now that 9.5.0 has reached general availability, the
matrix must test against the released `9.5.0` image instead of the snapshot build. Leaving the
snapshot entry in place means CI keeps exercising an unreleased/mutable build and keeps treating
that version's acceptance failures as non-blocking (`continue-on-error`), even though a GA release
should be held to the same blocking bar as every other released stack version in the matrix.

## What Changes

- Replace the `9.5.0-SNAPSHOT` matrix entry in `.github/workflows/provider.yml` with `9.5.0`, so the
  acceptance matrix builds and tests against the released image rather than the snapshot build.
- Because the promoted entry no longer matches `endsWith(matrix.version, '-SNAPSHOT')`, add `9.5.0`
  explicitly wherever that suffix match was standing in for per-version step gating (notably the
  Fleet setup step's explicit version list), so the promoted version keeps the Fleet coverage it had
  while still labeled as a snapshot.
- As a result, `9.5.0` acceptance-test failures become blocking (`continue-on-error` no longer
  applies) like every other non-snapshot matrix entry, and the snapshot-failure PR warning comment
  no longer applies to `9.5.0`.
- Add an explicit requirement to the `ci-build-lint-test` capability describing how a snapshot-to-GA
  version promotion must be carried out, so this same promotion is done consistently in future stack
  releases and doesn't silently drop per-version step coverage (see Design for the concrete gap this
  closes).
- Making `9.5.0` blocking surfaced three real acceptance regressions against the released Kibana
  9.5.0 GA image that `continue-on-error` had been masking under the snapshot label. Fixed in this
  same change (see Design for root causes) rather than deferred, since a follow-up PR would have
  needed to reopen the exact CI failures this promotion is meant to make visible:
  - A null-preservation bug in `PopulateFromAPI` across 10 dashboard panel types, where a
    "type-change recovery" branch fired on every same-type update (not just genuine type changes),
    skipping REQ-009 null-preservation whenever Kibana 9.5.0 GA returns a concrete default for a
    field the practitioner left unset.
  - Kibana 9.5.0 GA now echoes a `name` key in `data_source_json` `data_view_spec` payloads that
    earlier versions omitted, breaking apply-consistency for Lens by-value chart types.
  - Kibana 9.5.0 GA rejects `ml_anomaly_charts_config.severity_threshold` raw ranges whose `min`
    does not match one of five canonical band boundaries — a genuine Kibana-side API change between
    the pre-GA snapshot and GA release, not a provider defect. The now-permanently-failing
    `TestAccResourceDashboardMlAnomalyChartsRawRange` test was removed and the `kibana-dashboard`
    capability's `severity_threshold` requirement updated to match.

## Capabilities

### Modified Capabilities
- `ci-build-lint-test`: acceptance matrix must promote a snapshot-labeled version to its GA release
  entry without losing per-version step coverage that had been keyed off the snapshot suffix.
- `kibana-dashboard`: `ml_anomaly_charts_config.severity_threshold`'s raw `min`/`max` form is
  documented as an alternative numeric spelling of the five canonical severity bands, not a
  general-purpose custom range — Kibana's API rejects non-canonical `min` values with HTTP 400.

## Impact

- `.github/workflows/provider.yml`: `test` job matrix, Fleet setup step condition.
- Provider Go code: 20 non-test files across `internal/kibana/dashboard/` fixing the three
  regressions above (10 panel `model.go` files for null-preservation; 11 `alignment.go` call sites
  plus one shared `lenscommon` helper for the `data_source_json` `name` key; `mlanomalycharts`
  schema/test changes for `severity_threshold`), plus their corresponding test files.
- CI acceptance coverage: `9.5.0` acceptance failures start blocking the `test` job and
  `Test Validation` (previously non-blocking under the snapshot label).
