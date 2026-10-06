## ADDED Requirements

### Requirement: Schema (REQ-001)

The resource SHALL expose the following top-level attributes:

```hcl
resource "elasticstack_elasticsearch_index_settings" "example" {
  id    = <computed, string>                     # "<cluster_uuid>/<index_name>"
  index = <required, string, forces replacement> # name of the target (already concrete) Elasticsearch index

  # One optional, individually-typed attribute per internal/elasticsearch/index.DynamicSettingsKeys
  # entry, sourced from the shared getDynamicSettingAttributes() function, e.g.:
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
- The dynamic-setting attributes SHALL be produced by `getDynamicSettingAttributes()` (owned by `internal/elasticsearch/index/settings_keys.go`) and merged into this resource's schema map, so their names, types, and descriptions are identical to the equivalent attributes on `elasticstack_elasticsearch_index`.
- `settings_json` SHALL be optional, typed `jsontypes.Normalized`, and validated as a JSON object. At plan time, any top-level key of `settings_json` that is a literal, exact match of an entry in `internal/elasticsearch/index.StaticSettingsKeys` SHALL be rejected with a validation error. Keys that are not in `StaticSettingsKeys` — including keys not present in `AllSettingsKeys` at all — SHALL be permitted (permissive on unknown keys).
- `elasticsearch_connection` is injected by the provider scaffold and SHALL NOT be declared manually in the schema factory.
- At least one of the typed dynamic-setting attributes or `settings_json` SHALL be set; a configuration that sets none of them SHALL be rejected at plan time.

#### Scenario: Schema validation — index is required

- GIVEN a configuration that omits `index`
- WHEN `terraform validate` runs
- THEN Terraform SHALL emit a required-attribute error

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

On read, the resource SHALL retrieve the index's settings via the existing `GetIndex` helper and populate only the typed dynamic-setting attributes and `settings_json` keys that are present in the previously stored state (the declared subset). Settings returned by Elasticsearch that are not part of the declared subset SHALL be silently ignored and SHALL NOT be written to state and SHALL NOT cause drift. If the previously stored state is empty (e.g. immediately after `terraform import`), the resource SHALL populate all `DynamicSettingsKeys`-derived attributes and `settings_json` from the full API response as the initial declared subset.

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

The resource `id` SHALL follow the format `<cluster_uuid>/<index_name>`, matching `elasticstack_elasticsearch_index_mappings`. The resource SHALL support `terraform import` using the same ID format via `resource.ImportStatePassthroughID`. Each resource instance SHALL target exactly one concrete, already-resolved index name; the resource SHALL NOT expand a wildcard pattern to manage settings across multiple indices within a single instance's state. Callers needing to manage settings across many concrete indices SHALL use Terraform's own `for_each` over index names resolved outside this resource (for example via the `elasticstack_elasticsearch_indices` data source).

#### Scenario: Import

- GIVEN an existing index with a known `index.mapping.total_fields.limit` of `5000`
- WHEN the user runs `terraform import elasticstack_elasticsearch_index_settings.example <cluster_uuid>/<index_name>`
- THEN the resource SHALL be added to state with all `DynamicSettingsKeys`-derived attributes populated from the API response
- AND a subsequent `terraform plan` with a narrowed config (declaring only `mapping_total_fields_limit`) SHALL show a diff proposing to unset the other imported attributes
- AND `terraform apply` SHALL converge state to the declared subset

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
