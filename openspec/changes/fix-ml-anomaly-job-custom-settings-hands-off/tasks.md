## 1. Read path — stop mirroring the API value unconditionally

- [ ] 1.1 In `internal/elasticsearch/ml/anomalydetectionjob/models_tf.go`, in
  `(plan *TFModel) fromAPIModel`, capture `plan.CustomSettings` into a local
  `priorCustomSettings` before it is overwritten (it currently holds the plan value on
  the write/read-after-write path, or the prior state value on the plain read/refresh
  path — see design.md "Read: `fromAPIModel`").
- [ ] 1.2 Replace the current unconditional block:
  ```go
  if apiModel.CustomSettings != nil {
      customSettingsJSON, err := json.Marshal(apiModel.CustomSettings)
      ...
      plan.CustomSettings = jsontypes.NewNormalizedValue(string(customSettingsJSON))
  } else {
      plan.CustomSettings = jsontypes.NewNormalizedNull()
  }
  ```
  with the three-way branch from design.md: null stays null; an empty-object prior
  (`"{}"`, semantically) stays `"{}"` regardless of the API response; any other prior
  object is refreshed from the live API value (or forced to `"{}"` if the API returns
  nothing for an owned bag).
- [ ] 1.3 Add a small helper (e.g. `isEmptyJSONObject(v jsontypes.Normalized) bool`) that
  decodes the value's JSON into `map[string]any` and reports `len(...) == 0`, rather than
  string-comparing against the literal `"{}"` (formatting can vary while remaining
  semantically equal). Unit-test it directly in `models_api_test.go` or a new
  `models_tf_test.go` table test.

## 2. Write path — make `"{}"` reach the Update Job API

- [ ] 2.1 In `internal/elasticsearch/ml/anomalydetectionjob/models_api.go`, change
  `UpdateAPIModel.CustomSettings` from `map[string]any \`json:"custom_settings,omitempty"\`` to
  `json.RawMessage \`json:"custom_settings,omitempty"\`` (a marshaled `{}` is 2 bytes and
  is not empty for `omitempty` purposes on `json.RawMessage`, unlike an empty map).
- [ ] 2.2 In `BuildFromPlan`, after unmarshaling `plan.CustomSettings.ValueString()` into
  `customSettings map[string]any`, re-marshal it into `json.RawMessage` before assigning
  to `u.CustomSettings`, per design.md "Write: `BuildFromPlan` / wire encoding". Keep the
  existing `!plan.CustomSettings.Equal(state.CustomSettings) && !plan.CustomSettings.IsNull()`
  guard — it already correctly skips sending when the plan is null.
- [ ] 2.3 Confirm `APIModel.CustomSettings` (used by `toPutJobRequest` for create) is
  unaffected — it already flows through an explicit `json.Marshal` into
  `req.CustomSettings json.RawMessage` (`models_api.go:297-300`) and already sends `{}`
  correctly for an empty-but-non-nil map. No change needed there; add a regression test
  (task 4) rather than modifying working code.

## 3. Verify `"{}"` semantics against a live Elasticsearch cluster

- [ ] 3.1 Before or alongside the acceptance test in task 4, confirm that
  `POST _ml/anomaly_detectors/<job_id>/_update` with body `{"custom_settings":{}}`
  actually clears every existing key in `custom_settings` on the target job (per
  design.md "Open Questions" — the agreed path treats this as likely given Update Job's
  documented replace-not-merge semantics, but it has not been confirmed against a live
  stack). If it does **not** clear existing keys, revisit the `"{}"` wipe semantics in
  the delta spec before this change is considered ready to implement further.

## 4. Acceptance tests

- [ ] 4.1 Add a test that reproduces the reported bug: create a job with
  `custom_settings` omitted from config, inject a `custom_settings` value directly via
  the Elasticsearch Update Job API outside of Terraform (simulating Kibana), then run
  `terraform plan` and `terraform apply` and assert **no diff** and **no error** (the
  core regression test for this issue).
- [ ] 4.2 Add a test for the `"{}"` wipe: start from a job with a populated
  `custom_settings`, set `custom_settings = "{}"` in config, apply, and assert state
  shows `"{}"` and (per task 3's finding) that the server-side bag is actually cleared.
- [ ] 4.3 Add a test for re-ownership drift: set `custom_settings` to a real object,
  apply, then have Elasticsearch/Kibana add an extra key outside Terraform, and assert
  the next `terraform plan` shows a diff and the next `terraform apply` replaces the bag
  with exactly the configured object (extras dropped) — confirming the "any other
  object" row of the contract still holds and did not regress.
- [ ] 4.4 Confirm the existing `TestAccResourceAnomalyDetectionJobComprehensive` and
  `TestAccResourceAnomalyDetectionJobNullAndEmpty` tests (which already cover
  `custom_settings` set-and-updated and `custom_settings = null` respectively) still
  pass unmodified — this change must not alter behavior for configs that already set the
  attribute.

## 5. Build and verify

- [ ] 5.1 Run `make build` and confirm the provider compiles without errors.
- [ ] 5.2 Run the targeted acceptance tests for the anomaly detection job resource
  (requires a running Elasticsearch stack; see
  [`dev-docs/high-level/testing.md`](../../../dev-docs/high-level/testing.md)).
- [ ] 5.3 Run `go test ./internal/elasticsearch/ml/anomalydetectionjob/...` (unit tests
  only) to confirm `models_api_test.go` / any new table tests pass.

## 6. OpenSpec

- [ ] 6.1 Keep the delta spec
  `openspec/changes/fix-ml-anomaly-job-custom-settings-hands-off/specs/elasticsearch-ml-anomaly-detection-job/spec.md`
  aligned with the implementation as it lands.
- [ ] 6.2 After merge: sync the delta into
  `openspec/specs/elasticsearch-ml-anomaly-detection-job/spec.md` (updating the existing
  `custom_settings`-related requirement text in REQ-025-026/REQ-027-031 rather than
  duplicating it) and run `make check-openspec`.
