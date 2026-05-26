## Context

The Fleet `PUT /api/fleet/outputs/{id}` request body uses `*string` fields with `json:"...,omitempty"`, so a nil `ConfigYaml` pointer serializes to "field absent" and the Fleet API interprets that as "leave the stored value alone". The response, however, always includes `config_yaml` (set either to the previously stored value or an empty string). The provider's `populateFromAPI` unconditionally wrote that echoed value into state. Because the resource schema marks `config_yaml` as `Sensitive: true`, Terraform's post-apply consistency check fires on any null↔value or value↔different-value transition for the attribute and aborts the apply.

Two failure modes were observed:

1. Outputs that never had `config_yaml` configured. The API returns `"config_yaml": ""` and the provider wrote `""` into state, conflicting with the planned null on every in-place update unrelated to `config_yaml`.
2. Outputs whose `config_yaml` was set and then removed. The API preserves and echoes the previously stored non-empty value, so the provider wrote that non-empty value over the planned null.

## Decision

Three coordinated changes:

1. **Empty-string normalization on read.** `fromAPICommonFields` calls a new `configYamlFromAPI` helper that maps both `nil` and `""` to a null `NormalizedYamlValue`. This addresses failure mode (1) without altering plan/apply semantics for outputs whose `config_yaml` is genuinely set.
2. **Preserve user-removed null across the API echo, in both Update and Read.** `fromAPICommonFields` captures the existing model's `ConfigYaml` before overwriting and restores a null over any non-null API echo. Because `Update` calls `populateFromAPI` on `planModel` (whose `ConfigYaml` reflects the user's intent), and `Read` calls it on the prior state (which carries the previously preserved null), both paths converge on the user's intent. To avoid breaking import — where the prior model carries only the importer-populated identity fields — the preservation is gated on the existing model's required `name` field being non-null, which it is after any successful create / read / update.
3. **Semantic YAML equality.** Migrating the attribute to `customtypes.NormalizedYamlType` ensures the Plugin Framework compares plan and applied values structurally. This protects against future regressions where Fleet might re-emit normalized YAML (whitespace, key reordering) for a value the user did set, which would otherwise still trip the sensitive-attribute consistency check.

## Alternatives Considered

- **Send an explicit empty string to force-clear server-side.** Rejected: the generated Fleet API types use `omitempty`, so a non-nil pointer to `""` still serializes as field-absent. Regenerating the kbapi to drop `omitempty` would risk regressing issue #1067 (sending `null` is rejected by Fleet).
- **Preserve plan-null only in `Update`, not in `Read`.** Rejected: the post-apply auto-refresh that the testing framework (and `terraform refresh`) runs would immediately re-fetch the API echo, leaving state with the value the user just removed. The next plan would show a perpetual diff trying to set `config_yaml` back to null. Preserving in both paths converges on the user's intent and matches the documented Fleet limitation.
- **Make `config_yaml` `Computed` so plan picks up whatever Fleet returns.** Rejected: the attribute is a user input, not a server-derived field; making it computed would break drift detection and conflict with the existing `Optional: true` semantics.
- **Use a plan modifier to suppress the value→null transition silently.** Rejected: the user would have no way to see that their attempt to remove the attribute was being silently ignored. Honouring plan-null in state (which is what the user asked for) is more honest, even though it suppresses drift detection on a subsequently-edited server-side value.

## Risks / Trade-offs

- Once `config_yaml` has been removed via configuration, Fleet still holds the stored value server-side but the resource no longer surfaces external changes to that value as drift. This is an intentional concession: the user removed the attribute from configuration, so suppressing drift detection for it matches user intent. Re-introducing `config_yaml` in configuration and applying will overwrite the server-side value as normal.
- Import detection relies on the existing model's `name` field being null. Any future change that pre-populates `name` during import (e.g. via a custom `ImportState`) would silently turn off the import discriminator and could cause import to suppress the API value. The relevant invariant is documented in `fromAPICommonFields` and exercised by both the existing `TestAccResourceOutputElasticsearch` import step and the new `import populates config_yaml from the API` unit test.
- The `NormalizedYamlValue` type validates that the configured `config_yaml` parses as YAML. Existing configurations that pass anything other than valid YAML (e.g. an arbitrary string the user expected Fleet to round-trip) will now error at plan time. The attribute is documented as a YAML config so this is considered desirable, but it is technically a tightened contract worth calling out in the CHANGELOG.
