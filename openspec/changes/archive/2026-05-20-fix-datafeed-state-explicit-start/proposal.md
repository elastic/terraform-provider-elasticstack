## Why

`elasticstack_elasticsearch_ml_datafeed_state` fails with `Provider produced inconsistent result after apply` whenever a user supplies an explicit `start` (and, by extension, `end`). After starting the datafeed, the provider reads `running_state.search_interval.start_ms` from Elasticsearch and writes it back into the `start` attribute — but Elasticsearch routinely returns a different value than the one requested (the interval is snapped to the next bucket boundary, or to the timestamp of the first matched document). Terraform's framework rejects the apply because a known config value changed during apply.

This issue is tracked as [#2353](https://github.com/elastic/terraform-provider-elasticstack/issues/2353). The previous fix (PR #1563, REQ-018) only covers the case where `start` is omitted; the explicit-start path remains broken and forces users to fall back to `terraform_data` + `local-exec` to call the ES API directly.

## What Changes

- **BREAKING (schema-level):** `start` becomes a pure user-input attribute. It is no longer `Computed` and is no longer overwritten by reads. The `SetUnknownIfStateHasChanges` plan modifier and `UseStateForUnknown` plan modifier on `start` are removed.
- Add two new computed read-only attributes that surface Elasticsearch's view of the active search interval:
  - `effective_search_start` (RFC 3339): mirrors `running_state.search_interval.start_ms` for a `started` datafeed.
  - `effective_search_end` (RFC 3339): mirrors `running_state.search_interval.end_ms` for a `started` datafeed; null when `running_state.real_time_configured = true` or the datafeed is `stopped`.
- On read/create/update, populate the new attributes from the Get Datafeed Stats API. The user-supplied `start`/`end` values are preserved verbatim in state.
- Delete `set_unknown_if_state_has_changes.go` (only consumer was the `start` attribute).
- Update `resource-description.md` and any HCL examples to document the new attributes and the new semantics for `start`.

## Capabilities

### New Capabilities
None.

### Modified Capabilities

- `elasticsearch-ml-datafeed-state`:
  - Replace REQ-017 (Read — start and end from API) so that `start`/`end` are NOT overwritten by reads; instead, the new computed attributes `effective_search_start`/`effective_search_end` are populated from `search_interval`.
  - Remove REQ-018 (`SetUnknownIfStateHasChanges` on `start`) — no longer needed once `start` is not Computed.
  - Add a new requirement covering the new computed attributes (population, null-handling for stopped/real-time/missing running_state, and round-trip behavior).
  - Update the Schema block to drop `Computed` from `start`, drop the "unknown when state changes" annotation, and add `effective_search_start` / `effective_search_end`.

## Impact

- `internal/elasticsearch/ml/datafeed_state/schema.go` — drop `Computed` and plan modifiers from `start`; add two new computed `effective_search_*` attributes.
- `internal/elasticsearch/ml/datafeed_state/models.go` — extend `MLDatafeedStateData` with two new RFC3339 fields; rewrite `SetStartAndEndFromAPI` to populate them (and to leave `Start`/`End` untouched on read). Drop the post-loop unknown→null reconciliation for `Start`/`End`.
- `internal/elasticsearch/ml/datafeed_state/read.go`, `update.go` — minor adjustments so the new fields round-trip; ensure `effective_search_*` are set to null when the datafeed is stopped.
- `internal/elasticsearch/ml/datafeed_state/set_unknown_if_state_has_changes.go` — deleted.
- `internal/elasticsearch/ml/datafeed_state/issue_2353_acc_test.go` — flip from `ExpectError` to positive assertions verifying `start` is preserved and `effective_search_start` reports the ES-reported value.
- `internal/elasticsearch/ml/datafeed_state/acc_test.go` — extend coverage for the new attributes and the explicit-start path; verify the import path still works without `start`/`end` being Computed.
- `internal/elasticsearch/ml/datafeed_state/resource-description.md` and example configs — document the new attributes and behavior.
- State migration: dropping `Computed` from `start` is technically a schema-shape change. Existing state files store `start` as a string regardless; the behavior change should be benign on re-plan because the value previously stored is what the user supplied (preserved) or what ES reported (still a valid RFC3339 string). Verify via an upgrade smoke test.
- Closes [#2353](https://github.com/elastic/terraform-provider-elasticstack/issues/2353).
