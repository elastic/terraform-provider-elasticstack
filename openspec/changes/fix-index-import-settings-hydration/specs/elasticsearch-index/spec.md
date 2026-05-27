## ADDED Requirements

### Requirement: Import settings hydration (REQ-IMPORT-HYDRATE)

After a successful `terraform import`, the read that follows SHALL fully populate all individual index setting fields and provider-side operational defaults in state, so that a subsequent `terraform plan` produces no changes when the user's configuration matches the imported index.

The resource SHALL implement this via a private state flag (`"import_hydration"`) written during `ImportState` and consumed (then cleared) in the following `Read`:

1. During `ImportState`, in addition to setting `id` via `ImportStatePassthroughID`, the resource SHALL write the `"import_hydration"` key to provider private state.
2. During the `Read` that follows, if `"import_hydration"` is present in private state, the resource SHALL:
   a. Parse `settings_raw` (flat JSON map with `index.*`-prefixed keys) and populate every individual setting field (`number_of_replicas`, `refresh_interval`, `analysis_analyzer`, `analysis_tokenizer`, `analysis_char_filter`, `analysis_filter`, `analysis_normalizer`, all slowlog thresholds, all block settings, and all other attributes in the settings schema) into the model.
   b. Set provider-side operational defaults for attributes that have no Elasticsearch equivalent: `deletion_protection = true`, `wait_for_active_shards = "1"`, `master_timeout = "30s"`, `timeout = "30s"`.
   c. Clear the `"import_hydration"` private state key so subsequent reads are unaffected.
3. The deprecated `settings` block SHALL NOT be populated during import hydration.
4. The `use_existing` attribute SHALL NOT be populated during import hydration.
5. Regular read operations (not preceded by an import) SHALL NOT be affected by this mechanism.

#### Scenario: Import produces fully-hydrated state for configured settings

- **GIVEN** an Elasticsearch index exists with `number_of_replicas = 1`, `refresh_interval = "30s"`, and a custom analyzer defined in `analysis_analyzer`
- **WHEN** the user runs `terraform import` with the index's composite id
- **THEN** the saved state SHALL contain `number_of_replicas = 1`, `refresh_interval = "30s"`, and `analysis_analyzer` set to the JSON representation of the custom analyzer

#### Scenario: Plan is clean after import when config matches

- **GIVEN** the user's Terraform config specifies `number_of_replicas`, `refresh_interval`, and `analysis_analyzer` values that match the existing index
- **WHEN** `terraform import` completes and `terraform plan` runs
- **THEN** the plan SHALL show no changes

#### Scenario: Operational defaults populated after import

- **GIVEN** a `terraform import` has just completed
- **WHEN** the following `terraform plan` runs
- **THEN** `deletion_protection`, `wait_for_active_shards`, `master_timeout`, and `timeout` SHALL each have their default values in state (`true`, `"1"`, `"30s"`, `"30s"` respectively) and SHALL NOT appear as changes in the plan when the user's config omits or matches those defaults

#### Scenario: Regular read is unaffected

- **GIVEN** a resource managed by Terraform (not just imported) with `number_of_replicas = 2` in config and state
- **WHEN** `terraform plan` runs a refresh read
- **THEN** `number_of_replicas` SHALL remain `2` in state and no drift SHALL be introduced for settings not present in the user's config
