## 1. Schema changes

- [x] 1.1 In `internal/elasticsearch/ml/datafeed_state/schema.go`, drop `Computed: true` and the `PlanModifiers` block (`UseStateForUnknown`, `SetUnknownIfStateHasChanges`) from the `start` attribute, leaving it as `Optional` only.
- [x] 1.2 In `internal/elasticsearch/ml/datafeed_state/schema.go`, add a new `effective_search_start` attribute: `Computed: true`, `CustomType: timetypes.RFC3339Type{}`, markdown description explaining that it mirrors `running_state.search_interval.start_ms`.
- [x] 1.3 In `internal/elasticsearch/ml/datafeed_state/schema.go`, add a new `effective_search_end` attribute: `Computed: true`, `CustomType: timetypes.RFC3339Type{}`, markdown description explaining that it mirrors `running_state.search_interval.end_ms` and is null when the datafeed is stopped or `real_time_configured = true`.
- [x] 1.4 Confirm `end` schema attribute is `Optional` only (no `Computed`) — no change expected; document the assertion in the PR.

## 2. Model changes

- [x] 2.1 In `internal/elasticsearch/ml/datafeed_state/models.go`, add two fields to `MLDatafeedStateData`: `EffectiveSearchStart timetypes.RFC3339` (tfsdk:`effective_search_start`) and `EffectiveSearchEnd timetypes.RFC3339` (tfsdk:`effective_search_end`).
- [x] 2.2 Rewrite `SetStartAndEndFromAPI` to populate `EffectiveSearchStart` / `EffectiveSearchEnd` from `running_state.search_interval.{start_ms,end_ms}` instead of `Start`/`End`. Keep the timezone-preservation behavior (`timeInSameLocation` against the prior `Start`/`End` for display) for the new fields.
- [x] 2.3 In `SetStartAndEndFromAPI`, set `EffectiveSearchEnd` to `timetypes.NewRFC3339Null()` when `running_state.real_time_configured == true`.
- [x] 2.4 In `SetStartAndEndFromAPI`, set both `EffectiveSearchStart` and `EffectiveSearchEnd` to `timetypes.NewRFC3339Null()` for any non-`started` state, or when `running_state` / `running_state.search_interval` is nil.
- [x] 2.5 Remove the trailing "if d.Start.IsUnknown() { d.Start = null }" / "if d.End.IsUnknown() { d.End = null }" reconciliation block from `SetStartAndEndFromAPI` — no longer applicable now that `Start`/`End` are not touched on read.

## 3. Resource update flow

- [x] 3.1 In `internal/elasticsearch/ml/datafeed_state/update.go::updateAfterMissedTransition`, remove the `if data.Start.IsUnknown() { data.Start = null }` block (no longer applicable). Set `data.EffectiveSearchStart` and `data.EffectiveSearchEnd` to null instead.
- [x] 3.2 In `internal/elasticsearch/ml/datafeed_state/read.go::readMLDatafeedState`, ensure `effective_search_start` / `effective_search_end` are initialised before calling `SetStartAndEndFromAPI` (default to null) so they are well-defined when the datafeed is stopped.

## 4. Remove the obsolete plan modifier

- [x] 4.1 Delete `internal/elasticsearch/ml/datafeed_state/set_unknown_if_state_has_changes.go`.
- [x] 4.2 Remove the `SetUnknownIfStateHasChanges()` reference from `schema.go` (already covered in task 1.1; re-verify the file compiles with no remaining import for it).

## 5. Documentation & examples

- [x] 5.1 Update `internal/elasticsearch/ml/datafeed_state/resource-description.md` to:
  - Note that `start` and `end` are user inputs that are preserved verbatim in state.
  - Describe the new computed `effective_search_start` and `effective_search_end` attributes.
  - Cross-reference issue #2353 in the changelog/notes if the template supports it.
- [x] 5.2 Update or add HCL examples under `examples/resources/elasticstack_elasticsearch_ml_datafeed_state/` (if present) to demonstrate using `effective_search_start` via an output.
- [x] 5.3 Run `make docs-generate` (or equivalent) to regenerate `docs/resources/elasticsearch_ml_datafeed_state.md`.
- [x] 5.4 Add a CHANGELOG entry under "Fixed" referencing issue #2353 and the new attributes, including a brief note about the one-time plan diff for existing state with explicit `start`.

## 6. Tests

- [x] 6.1 Update `internal/elasticsearch/ml/datafeed_state/issue_2353_acc_test.go`: flip from `ExpectError: regexp.MustCompile(...)` to positive assertions on `start` (= configured value) and `effective_search_start` (= ES-reported value, e.g. `2022-01-01T00:10:00Z`). Rename the test (e.g. `TestAccResourceMLDatafeedState_explicitStartPreserved`) and move into `acc_test.go` if appropriate, or keep as a dedicated regression test referencing the issue.
- [x] 6.2 Extend `internal/elasticsearch/ml/datafeed_state/acc_test.go` with a test covering an explicit `start` round-trip across plan→apply→plan (verifying no drift on the second plan).
- [x] 6.3 Extend `acc_test.go` with a test covering an explicit `end` round-trip (mirror of 6.2, ensuring `end` is not rewritten either).
- [x] 6.4 Extend `acc_test.go` import test (`TestAccResourceMLDatafeedState_import`) to assert that `effective_search_start` / `effective_search_end` are populated after import, and that `start` remains null after import when not in config.
- [x] 6.5 Add a unit test for `SetStartAndEndFromAPI` exercising: (a) started + search_interval present → populates effective fields, leaves Start/End untouched; (b) started + real_time_configured → effective_end = null; (c) stopped → both effective fields = null; (d) started + nil running_state → both effective fields = null.
- [x] 6.6 Add a unit test that confirms `MLDatafeedStateData` correctly serializes the new `effective_search_*` tfsdk tags via the framework's reflection.

## 7. Verification

- [x] 7.1 Run `make build` and ensure the provider compiles.
- [x] 7.2 Run targeted unit tests: `go test ./internal/elasticsearch/ml/datafeed_state/...`.
- [x] 7.3 Run targeted acceptance tests (requires Elastic Stack — see `dev-docs/high-level/testing.md`): `TF_ACC=1 go test ./internal/elasticsearch/ml/datafeed_state/... -run 'TestAccResourceMLDatafeedState|TestAccReproduceIssue2353|TestAccResourceMLDatafeedState_explicitStartPreserved'`.
- [x] 7.4 Run `make check-openspec` to confirm the spec delta validates.
- [x] 7.5 Run `make check-lint` (covers OpenSpec, gofmt, etc.).
- [x] 7.6 Manual smoke test: apply a config with `bucket_span = "15m"` and `start = "<not-on-boundary>"`, confirm no inconsistency error and that `effective_search_start` shows the bucket-aligned value (mirrors the reporter's scenario in #2353).
