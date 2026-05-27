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
- **Private state flag** (chosen): `ImportState` writes `"import_hydration"` to private state; `Read` checks for it, passes `hydrateAll=true` down the call stack, populates before the single `resp.State.Set`, then clears the flag. Precisely scoped, no regressions, follows the existing `sort_config` pattern.

### D2: `hydrateAll bool` flows through `readIndex` → `populateFromAPI`

Rather than checking the private state flag in `Read` and calling hydration after `resp.State.Set`, the flag is resolved once in `Read` and passed as a parameter so hydration happens inside `populateFromAPI` before the model is returned. This keeps all model-population logic within `populateFromAPI` and avoids a second `resp.State.Set` call.

### D3: `hydrateAllSettingsFromRaw` parses `settings_raw` (not the `IndexSettings` struct directly)

`settings_raw` is already the correct flat JSON representation. Parsing it is consistent with the existing `populateSortFromSettings` / `extractSortSetting` pattern in `sort_read.go`. The alternative (using the typed `IndexSettings.Index.*` struct fields, which requires removing `FlatSettings(true)`) would change the `settings_raw` format and have wide blast radius across all index read operations — better left as a future refactor.

### D4: Analysis settings handled as a nested JSON object under `"index.analysis"`

With `FlatSettings(true)`, Elasticsearch returns analysis settings as a single nested JSON object under the flat key `"index.analysis"` (not further flattened). `hydrateAllSettingsFromRaw` looks up this key and extracts sub-keys `analyzer`, `tokenizer`, `char_filter`, `filter`, `normalizer`, marshalling each to `jsontypes.Normalized` for the corresponding model fields.

### D5: New file `settings_read.go` alongside `sort_read.go`

`hydrateAllSettingsFromRaw` and `populateOperationalDefaults` live in a new `settings_read.go` following the exact pattern of `sort_read.go`. This keeps read-path helpers co-located and separate from the main model file.

## Risks / Trade-offs

- **Type conversion edge cases** → The flat settings map returns all scalar values as JSON strings (e.g., `"1"` for integers, `"true"` for booleans). Conversion uses `strconv.ParseInt` and `strconv.ParseBool`. An unexpected non-numeric/non-bool string from a future ES version would produce a diagnostic error on import. Mitigation: skip fields that fail conversion rather than erroring, and log the issue — consistent with how the provider handles unexpected values in other places.
- **Re-import** → If a resource is re-imported, the flag is set again and hydration runs again. This is benign — it re-hydrates all settings from the current ES state.
- **Private state key lifecycle** → The `"import_hydration"` key exists in private state for exactly one read cycle (set in `ImportState`, cleared in the following `Read`). If `Read` errors before clearing the flag, the key persists and hydration retries on the next read — also benign.

## Migration Plan

No migration required. This is a purely additive change to the import path. Existing managed resources are unaffected. No state upgrades, no schema version bumps.

## Open Questions

_(none — design is fully resolved)_
