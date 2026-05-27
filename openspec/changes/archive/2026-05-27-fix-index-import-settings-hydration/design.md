## Context

The `elasticstack_elasticsearch_index` resource uses Terraform Plugin Framework. Its `Read` method calls `readIndex`, which calls `populateFromAPI` on the current state model. `populateFromAPI` populates `name`, `concrete_name`, `alias`, `mappings`, and (via `setSettingsFromAPI`) `settings_raw`. It does not populate individual setting fields (`number_of_replicas`, `refresh_interval`, analysis attributes, etc.) because those fields are `Optional` (not `Computed`) — the resource only tracks them when the user explicitly configures them.

This is correct for regular read/plan/apply cycles: after `apply`, those fields are already in state from the plan. After `terraform import`, however, only `id` is in state (set by `ImportStatePassthroughID`). The following `Read` call leaves all individual fields null, causing drift against any configured values.

The Elasticsearch Get Indices API is called with `FlatSettings(true)`, which returns all settings as dot-prefixed flat strings (`index.number_of_replicas`, `index.refresh_interval`, etc.). The `go-elasticsearch` `IndexSettings.UnmarshalJSON` stores unrecognised flat keys in a catch-all `map[string]json.RawMessage`; `MarshalJSON` inlines them back. This is what produces `settings_raw`. The data needed for hydration is therefore already present in `settings_raw` — it just needs to be parsed into the individual model fields.

Private state is already used by this resource for `sort_config` (written by `saveSortConfig` in `Read`, consumed by `sortMigrationPlanModifier` during planning). The `ImportStateResponse.Private` field is pre-initialised to `EmptyProviderData` by the framework before `ImportState` is called, so writing to it is safe without any additional initialisation.

## Goals / Non-Goals

**Goals:**
- After `terraform import`, a subsequent `terraform plan` produces no changes when the user's config matches the imported index.
- Individual settings fields, analysis attributes, and provider-side operational defaults are populated in state after import.
- Regular read/plan/apply cycles are completely unaffected.

**Non-Goals:**
- Full drift detection for settings not in the user's config (that would require always-populating, which is a separate behaviour change and out of scope).
- Changing the `settings_raw` format (remains flat with `index.*` prefix).
- Populating the deprecated `settings { setting { } }` block on import.
- Populating `use_existing` (not meaningful on import).

## Decisions

### D1: Import-scoped hydration via private state flag (over always-populate or two-pass Read)

Three approaches were considered:
- **Always-populate** (`setSettingsFromAPI` always writes individual fields): Fixes import but causes regression — users who don't configure `number_of_replicas` would suddenly see it appear in state after a regular refresh, showing drift.
- **Two-pass Read** (hydrate after `resp.State.Set`, then `Set` again): Works mechanically but is inelegant — two state writes per import read cycle.
- **Private state flag** (chosen): `ImportState` writes `"import_hydration"` to private state; `Read` checks for it, passes `hydrateAll=true` down the call stack, and populates all settings before the single `resp.State.Set` — but does NOT clear the flag. `ModifyPlan` then prunes hydrated values for Optional-only settings not in the user's config, and finally clears the flag. Precisely scoped, no regressions, follows the existing `sort_config` pattern.

### D2: `hydrateAll bool` flows through `readIndex` → `populateFromAPI`

Rather than checking the private state flag in `Read` and calling hydration after `resp.State.Set`, the flag is resolved once in `Read` and passed as a parameter so hydration happens inside `populateFromAPI` before the model is returned. This keeps all model-population logic within `populateFromAPI` and avoids a second `resp.State.Set` call.

### D3: `hydrateAllSettingsFromRaw` parses `settings_raw` (not the `IndexSettings` struct directly)

`settings_raw` is already the correct flat JSON representation. Parsing it is consistent with the existing `populateSortFromSettings` / `extractSortSetting` pattern in `sort_read.go`. The alternative (using the typed `IndexSettings.Index.*` struct fields, which requires removing `FlatSettings(true)`) would change the `settings_raw` format and have wide blast radius across all index read operations — better left as a future refactor.

### D4: Analysis settings handled as a nested JSON object under `"index.analysis"`

With `FlatSettings(true)`, Elasticsearch returns analysis settings as a single nested JSON object under the flat key `"index.analysis"` (not further flattened). `hydrateAllSettingsFromRaw` looks up this key and extracts sub-keys `analyzer`, `tokenizer`, `char_filter`, `filter`, `normalizer`, marshalling each to `jsontypes.Normalized` for the corresponding model fields.

### D5: New file `settings_read.go` alongside `sort_read.go`

`hydrateAllSettingsFromRaw` and `populateOperationalDefaults` live in a new `settings_read.go` following the exact pattern of `sort_read.go`. This keeps read-path helpers co-located and separate from the main model file.

### D6: `ModifyPlan` prunes unconfigured Optional-only settings to prevent post-import drift

Individual settings attributes (`number_of_replicas`, `refresh_interval`, etc.) are `Optional` (not `Computed`) in the schema. If `Read` hydrates them from ES, the plan comparison sees state=<ES value> vs config=null for any setting the user did not configure, producing unwanted drift.

`ModifyPlan` (resource-level, `resource.ResourceWithModifyPlan`) resolves this: it checks for the `"import_hydration"` flag and, for each Optional-only settings field where the config value is null, overrides the planned state value to null. Only the fields the user actually configured survive into the applied state.

The provider-side operational defaults (`deletion_protection`, `wait_for_active_shards`, `master_timeout`, `timeout`) are `Optional+Computed` — the plan keeps prior-state values for unset Computed fields — so they are unaffected by this pruning step.

## Risks / Trade-offs

- **Type conversion edge cases** → The flat settings map stores all scalar values as JSON strings (raw bytes for `number_of_replicas = 1` are `"1"` including the surrounding quotes). Conversion first `json.Unmarshal`s the raw bytes into a Go `string`, then applies `strconv.ParseInt` for `types.Int64` fields or `strconv.ParseBool` for `types.Bool` fields. A non-numeric/non-bool string from a future ES version would leave the field unset. Mitigation: skip fields that fail conversion, consistent with how the provider handles unexpected values elsewhere.
- **Re-import** → If a resource is re-imported, the flag is set again, hydration runs again, and `ModifyPlan` prunes unconfigured settings again. This is benign.
- **Private state key lifecycle** → The `"import_hydration"` key exists in private state for one read+plan cycle: set in `ImportState`, read in `Read` to trigger hydration (not cleared there), read and cleared in the following `ModifyPlan`. If `Read` errors, the key persists and hydration retries on the next read — benign. If `ModifyPlan` errors before clearing, the key persists and the pruning retries on the next plan — also benign.

## Migration Plan

No migration required. This is a purely additive change to the import path. Existing managed resources are unaffected. No state upgrades, no schema version bumps.

## Open Questions

_(none — design is fully resolved)_
