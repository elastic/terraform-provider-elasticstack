## ADDED Requirements

### Requirement: Schema (REQ-001)

The resource SHALL expose the following top-level attributes:

```hcl
resource "elasticstack_elasticsearch_index_settings" "example" {
  id    = <computed, string>                     # "<cluster_uuid>/<index_name>"
  index = <required, string, forces replacement> # name of the target (already concrete) Elasticsearch index

  # One optional, individually-typed attribute per internal/elasticsearch/index.DynamicSettingsKeys
  # entry, sourced from the shared GetDynamicSettingAttributes() function, e.g.:
  mapping_total_fields_limit = <optional, int64>
  number_of_replicas         = <optional, int64>
  refresh_interval           = <optional, string>
  # ... remaining DynamicSettingsKeys-derived attributes, same names/types as elasticstack_elasticsearch_index

  settings_json = <optional, JSON string>         # escape hatch for dynamic settings not yet typed

  elasticsearch_connection { ... }                 # standard provider connection block
}
```

- `id` SHALL be computed and unknown until create completes; it SHALL use `stringplanmodifier.UseStateForUnknown()`.
- `index` SHALL be required and SHALL force resource replacement when changed. It SHALL be treated as an already-resolved concrete index name; the resource SHALL NOT perform date-math resolution.
- The dynamic-setting attributes SHALL be produced by `GetDynamicSettingAttributes()` (owned by `internal/elasticsearch/index/settings_keys.go`) and merged into this resource's schema map, so their names, types, and descriptions are identical to the equivalent attributes on `elasticstack_elasticsearch_index`.
- `settings_json` SHALL be optional, typed `jsontypes.Normalized`, and validated as a non-empty JSON object (`{}` is rejected at plan time, since it declares no settings). At plan time, any top-level key of `settings_json` that is a literal, exact match of an entry in `internal/elasticsearch/index.StaticSettingsKeys` SHALL be rejected with a validation error. Keys that are not in `StaticSettingsKeys` — including keys not present in `AllSettingsKeys` at all — SHALL be permitted (permissive on unknown keys).
- `settings_json` SHALL use flat dotted setting keys only (e.g. `"index.max_result_window"` or `"max_result_window"`), matching the flat form returned by `GetIndex`. Nested object values SHALL be rejected at plan time (e.g. `{"index": {"number_of_replicas": 2}}`), so each top-level key is a complete setting path shared by overlap validation, diffing and read. Explicit JSON `null` values SHALL be rejected at plan time; to reset a setting, the user omits it (REQ-003).
- A key set via a typed dynamic-setting attribute SHALL NOT also appear in `settings_json`. At plan time, after canonicalizing key spellings (with or without the `index.` prefix), any overlapping key SHALL be rejected with a validation error.
- `elasticsearch_connection` is injected by the provider scaffold and SHALL NOT be declared manually in the schema factory.
- At least one of the typed dynamic-setting attributes or `settings_json` SHALL be set; a configuration that sets none of them SHALL be rejected at plan time.

#### Scenario: Schema validation — index is required

- GIVEN a configuration that omits `index`
- WHEN `terraform validate` runs
- THEN Terraform SHALL emit a required-attribute error

#### Scenario: Schema validation — typed attribute and settings_json overlap

- GIVEN `number_of_replicas = 1` and `settings_json = jsonencode({ "number_of_replicas" = 2 })`
- WHEN `terraform validate` or `terraform plan` runs
- THEN Terraform SHALL emit a validation error identifying the overlapping key

#### Scenario: Schema validation — settings_json rejects nested objects

- GIVEN `number_of_replicas = 1` and `settings_json = jsonencode({ index = { number_of_replicas = 2 } })`
- WHEN `terraform validate` or `terraform plan` runs
- THEN Terraform SHALL emit a validation error stating `settings_json` must use flat dotted keys and not nested objects

#### Scenario: Schema validation — settings_json rejects explicit null

- GIVEN `settings_json = jsonencode({ refresh_interval = null, max_result_window = 20000 })`
- WHEN `terraform validate` or `terraform plan` runs
- THEN Terraform SHALL emit a validation error stating that `null` is not allowed and that a setting is reset by omitting it

#### Scenario: Schema validation — settings_json rejects an empty object

- GIVEN `settings_json = jsonencode({})`
- WHEN `terraform validate` or `terraform plan` runs
- THEN Terraform SHALL emit a validation error stating `settings_json` must declare at least one setting

#### Scenario: Schema validation — settings_json rejects a static key

- GIVEN `settings_json = jsonencode({ "number_of_shards" = 3 })`
- WHEN `terraform validate` or `terraform plan` runs
- THEN Terraform SHALL emit an attribute validation error on `settings_json` stating that `number_of_shards` can only be set at index creation time
- AND no API call SHALL be issued

#### Scenario: Schema validation — settings_json permits an unmodeled dynamic key

- GIVEN `settings_json = jsonencode({ "some.future.dynamic.setting" = "value" })`, where the key is not present in `AllSettingsKeys`
- WHEN `terraform validate` runs
- THEN Terraform SHALL NOT emit a validation error for that key

#### Scenario: Schema validation — at least one setting must be declared

- GIVEN a configuration that sets `index` only, with no typed dynamic-setting attribute and no `settings_json`
- WHEN `terraform validate` or `terraform plan` runs
- THEN Terraform SHALL emit a validation error stating at least one setting must be declared

### Requirement: Create — index must exist (REQ-002)

On create, the resource SHALL verify that the target index exists (via the existing `GetIndex` helper) before issuing `PUT /{index}/_settings`. If the index does not exist, the resource SHALL return an error diagnostic and SHALL NOT create any Elasticsearch resource or compute an `id`.

#### Scenario: Create fails when index is absent

- GIVEN the target `index` does not exist in Elasticsearch
- WHEN `terraform apply` runs the create operation
- THEN Terraform diagnostics SHALL include an error stating the index was not found
- AND no `PUT /{index}/_settings` call SHALL be issued

#### Scenario: Create succeeds when index exists

- GIVEN the target `index` exists in Elasticsearch
- AND the configuration declares `mapping_total_fields_limit = 5000`
- WHEN `terraform apply` runs the create operation
- THEN `PUT /{index}/_settings` SHALL be called with `{"index": {"mapping": {"total_fields": {"limit": 5000}}}}` (or the flat-key equivalent) as the request body
- AND the resource SHALL be added to state with `id = "<cluster_uuid>/<index_name>"`

### Requirement: Update — declared settings diff, null resets (REQ-003)

On update, the resource SHALL compute the union of declared typed dynamic-setting attributes and declared `settings_json` keys for both plan and prior state, and SHALL call `PUT /{index}/_settings` only when that union differs between plan and state. Settings present in prior state but absent from the plan (removed from configuration) SHALL be explicitly included in the update payload with a `null` value, so Elasticsearch resets them to their default — consistent with `elasticstack_elasticsearch_index`'s existing `updateSettings` behavior in `internal/elasticsearch/index/index/update.go`. No existence check is required on update.

#### Scenario: Changed setting triggers update

- GIVEN a managed resource with `mapping_total_fields_limit = 2300` in state
- WHEN the user changes the configuration to `mapping_total_fields_limit = 5000` and runs `terraform apply`
- THEN `PUT /{index}/_settings` SHALL be called with `{"index.mapping.total_fields.limit": 5000}` (or nested equivalent)
- AND the next `terraform plan` SHALL show no diff

#### Scenario: Removing a declared setting resets it via null

- GIVEN a managed resource with `mapping_total_fields_limit = 5000` in state
- WHEN the user removes the `mapping_total_fields_limit` attribute from configuration entirely and runs `terraform apply`
- THEN `PUT /{index}/_settings` SHALL be called with `{"index.mapping.total_fields.limit": null}`
- AND the attribute SHALL no longer appear in the resource's declared-subset state after the next read

#### Scenario: No-op update when nothing changed

- GIVEN a managed resource whose plan is identical to its prior state
- WHEN `terraform apply` runs
- THEN no `PUT /{index}/_settings` call SHALL be issued

### Requirement: Read — declared subset only (REQ-004)

On read, the resource SHALL retrieve the index's settings via the existing `GetIndex` helper and populate only the typed dynamic-setting attributes and `settings_json` keys that are present in the previously stored state (the declared subset). Settings returned by Elasticsearch that are not part of the declared subset SHALL be silently ignored and SHALL NOT be written to state and SHALL NOT cause drift. Full hydration SHALL occur only on the first read after `terraform import`, signalled by an import-specific private-state key set by the resource's `ImportState` implementation (alongside `id` and `index`) and cleared by that read. An empty set of tracked settings SHALL NOT be treated as an import. On that import read, the resource SHALL populate the known `DynamicSettingsKeys`-derived typed attributes from the API response as the initial declared subset and SHALL leave `settings_json` unset, so that static and metadata settings (e.g. `index.number_of_shards`, `index.uuid`) are never adopted into state.

When populating `settings_json` keys on read, the resource SHALL reconcile each API value (returned as a string because `GetIndex` requests flat settings) with the scalar type declared in state, converting numeric and boolean strings back to JSON numbers and booleans in the same way as the existing reader in `internal/elasticsearch/index/index/settings_read.go`, so that an unchanged `settings_json` does not produce false drift.

Ownership of a setting is defined by its presence in state (a non-null typed attribute or a `settings_json` key). When Elasticsearch no longer reports a tracked setting (for example because it was reset outside Terraform), read SHALL set that typed attribute to null or drop that key from `settings_json`, so the drift is shown; configuration that still declares it SHALL cause the plan to set it again. Outside the import read, read SHALL NOT add any setting to state, even when no tracked settings remain.

#### Scenario: Unrelated settings do not cause drift

- GIVEN a resource that declares only `mapping_total_fields_limit`
- AND the index also has `index.number_of_replicas` set to a value never declared by this resource
- WHEN `terraform plan` runs
- THEN the plan SHALL show no diff
- AND `number_of_replicas` SHALL NOT appear in this resource's state

#### Scenario: Declared setting changed out-of-band surfaces as drift

- GIVEN a resource that declares `mapping_total_fields_limit = 5000` in state
- AND the setting is changed to `2300` directly via the Elasticsearch API (outside Terraform)
- WHEN `terraform plan` runs
- THEN the plan SHALL show a diff proposing to change `mapping_total_fields_limit` back to `5000`

#### Scenario: Import hydrates once via the private-state flag

- GIVEN an existing index with several dynamic settings, imported via `terraform import`
- WHEN the first read runs
- THEN the known dynamic typed attributes SHALL be populated from the API response and `settings_json` SHALL be unset
- AND the import flag SHALL be cleared, so later reads do not hydrate

#### Scenario: Repeated refresh after external reset does not adopt settings

- GIVEN a resource that declares only `number_of_replicas = 2`, and the index also has `mapping_total_fields_limit` set to a value never declared by this resource
- AND `number_of_replicas` is reset outside Terraform so Elasticsearch no longer reports it
- WHEN `terraform refresh` runs twice
- THEN the first refresh SHALL set `number_of_replicas` to null in state, showing drift
- AND the second refresh SHALL NOT adopt `mapping_total_fields_limit` or any other setting
- AND `terraform apply` SHALL set `number_of_replicas = 2` again without sending `null` for any other setting

#### Scenario: Scalar types round-trip without drift

- GIVEN `settings_json = jsonencode({ "index.max_result_window" = 20000, "index.blocks.read_only" = false })` has been applied
- AND Elasticsearch returns `"20000"` and `"false"` as strings in the flat settings response
- WHEN `terraform plan` runs
- THEN the plan SHALL show no diff

#### Scenario: Not found on read removes from state

- GIVEN the target index is deleted outside Terraform
- WHEN `terraform refresh` or `terraform plan` runs
- THEN the resource SHALL be removed from state
- AND Terraform SHALL propose recreating it on the next apply

### Requirement: Delete — no-op (REQ-005)

On `terraform destroy`, the resource SHALL remove itself from Terraform state without issuing any API call. Settings are not reset or reverted on destroy. This is a fixed behavior in v1; there is no configurable destroy option.

The resource description and documentation SHALL clearly state that `destroy` does not revert or clear settings on the index.

#### Scenario: Destroy leaves index settings intact

- GIVEN a managed resource with one or more declared settings
- WHEN `terraform destroy` runs
- THEN no Elasticsearch API call SHALL be issued
- AND the resource SHALL be removed from Terraform state
- AND the settings SHALL remain on the Elasticsearch index unchanged

### Requirement: Identity, import, and scope to one concrete index (REQ-006)

The resource `id` SHALL follow the format `<cluster_uuid>/<index_name>`, matching `elasticstack_elasticsearch_index_mappings`. The resource SHALL support `terraform import` using the same ID format via a custom `ImportState` that parses the composite ID, sets `id` and `index`, and sets the import private-state key used by REQ-004 (plain `resource.ImportStatePassthroughID` is not sufficient, as it sets only `id`). Each resource instance SHALL target exactly one concrete, already-resolved index name; the resource SHALL NOT expand a wildcard pattern to manage settings across multiple indices within a single instance's state. Callers needing to manage settings across many concrete indices SHALL use Terraform's own `for_each` over index names resolved outside this resource (for example via the `elasticstack_elasticsearch_indices` data source).

#### Scenario: Import

- GIVEN an existing index with a known `index.mapping.total_fields.limit` of `5000`
- WHEN the user runs `terraform import elasticstack_elasticsearch_index_settings.example <cluster_uuid>/<index_name>`
- THEN the resource SHALL be added to state with the known `DynamicSettingsKeys`-derived typed attributes populated from the API response
- AND `settings_json` SHALL be unset
- AND a subsequent `terraform plan` with a narrowed config (declaring only `mapping_total_fields_limit`) SHALL show a diff proposing to unset the other imported typed attributes
- AND `terraform apply` SHALL succeed, sending only dynamic settings as `null`, and converge state to the declared subset

#### Scenario: for_each over multiple concrete indices

- GIVEN a list of concrete index names resolved outside this resource
- WHEN the user declares `resource "elasticstack_elasticsearch_index_settings" "x" { for_each = toset(local.indices) ... }`
- THEN each resource instance SHALL independently manage settings for its own resolved `index` value
- AND no instance SHALL affect the settings of any other instance's index

### Requirement: API errors surface as diagnostics (REQ-007)

When the Elasticsearch API returns a non-success response (other than 404 on read), the resource SHALL surface the API error to Terraform diagnostics rather than silently ignoring it.

#### Scenario: API failure on update

- GIVEN a valid configuration change to a declared setting
- WHEN `PUT /{index}/_settings` returns a non-2xx response (e.g. the target index was closed)
- THEN Terraform diagnostics SHALL include an error with the API's error detail
- AND the resource state SHALL NOT be updated to reflect the failed change
