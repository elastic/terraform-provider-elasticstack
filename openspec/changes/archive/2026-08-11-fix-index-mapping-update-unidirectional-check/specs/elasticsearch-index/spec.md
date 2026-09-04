## MODIFIED Requirements

### Requirement: Update flow (REQ-015–REQ-018)

On update, the resource SHALL only call the relevant update APIs when the corresponding values have changed. Alias changes SHALL be applied by deleting aliases removed from config (via Delete Alias API) and upserting all aliases present in plan (via Put Alias API). Dynamic setting changes SHALL be applied by calling the Put Settings API with the diff, setting removed dynamic settings to `null` in the request. All update APIs SHALL target the persisted concrete index identity from state / `id`, not the configured `name`. After all updates, the resource SHALL perform a read to refresh state while preserving any configured `name` already stored in state.

Mapping changes SHALL be applied using a **unidirectional** decision: the resource SHALL call the Put Mapping API whenever the planned `mappings` contain field or top-level mapping content that is not already present in the prior state `mappings`. The resource SHALL NOT call the Put Mapping API when the prior state `mappings` are a non-drifting superset of the planned `mappings` (state already contains everything the plan specifies, including Elasticsearch- or template-injected extras such as additional `properties`, `dynamic_templates`, or `_meta`). This decision SHALL NOT use the bidirectional semantic-equality check that also treats "plan is a superset of state" as equal — that bidirectional form remains reserved for plan-time semantic equality and replacement decisions (see the "Mappings plan modifier and semantic equality" requirement) and MUST NOT gate whether the Put Mapping API is called.

#### Scenario: Removed alias is deleted

- WHEN state has alias `old_alias` and config does not
- THEN update SHALL call the Delete Alias API for `old_alias`

#### Scenario: Removed dynamic setting is nulled

- WHEN state has a dynamic setting value and config removes it
- WHEN update runs
- THEN the resource SHALL send that setting as `null` in the Put Settings request

#### Scenario: Template-injected mappings do not cause mapping update

- **GIVEN** an index is created with user-owned `mappings`
- **AND** a matching index template injects additional mapping `properties`, `dynamic_templates`, or other top-level mapping keys
- **WHEN** Terraform refreshes and plans the same index configuration
- **THEN** the resource SHALL treat the template-injected mapping content as non-drift and SHALL NOT call the Put Mapping API solely for those template-owned differences

#### Scenario: Adding a mapping field calls the Put Mapping API

- **GIVEN** an index is managed with user-owned `mappings` and no configuration change other than a new field added to `mappings.properties`
- **WHEN** update runs
- **THEN** the resource SHALL call the Put Mapping API with the planned `mappings`
- **AND** the new field SHALL be present in the live cluster's mapping after apply

#### Scenario: State already covering plan skips the Put Mapping API

- **GIVEN** the prior state `mappings` already contain every field and top-level key present in the planned `mappings` (for example, only Elasticsearch- or template-injected extras differ)
- **WHEN** update runs
- **THEN** the resource SHALL NOT call the Put Mapping API

### Requirement: Opt-in adoption of existing indices via `use_existing`

The set of static settings compared during `use_existing` adoption SHALL be extended to include `sort.missing` and `sort.mode`. When these settings are explicitly set in the plan, the adoption flow SHALL compare them against the existing index's static settings and SHALL return an error diagnostic when they differ, consistent with the behavior for `sort.field` and `sort.order`.

When adopting an existing index, mapping reconciliation between the plan and the existing index's mappings SHALL use the same unidirectional Put Mapping decision defined in the "Update flow" requirement: the adoption flow SHALL call the Put Mapping API when the planned `mappings` contain content not already present in the existing index's mappings, and SHALL NOT call it when the existing index's mappings already cover everything the plan specifies. Fields present in the existing index's mappings but omitted from the plan's `mappings` SHALL be retained (the Put Mapping API cannot remove fields) and SHALL NOT be treated as a mismatch requiring an error diagnostic.

#### Scenario: Adoption compares `sort.missing` against existing index

- **GIVEN** `use_existing = true` and an existing index where `index.sort.missing` is `["_last"]`
- **AND** the plan specifies `sort = [{ field = "date", missing = "_first" }]`
- **WHEN** create runs
- **THEN** the adoption flow SHALL return an error diagnostic naming the mismatched `sort.missing` value
- **AND** SHALL NOT call any mutating API on the index

#### Scenario: Adoption writes a field the plan adds and retains a field the plan omits

- **GIVEN** `use_existing = true` and an existing index whose live mapping already has field `legacy_field` and does not have field `new_field`
- **AND** the plan's `mappings` specifies `new_field` and does not specify `legacy_field`
- **WHEN** create runs and adopts the existing index
- **THEN** the adoption flow SHALL call the Put Mapping API so that `new_field` is present in the live cluster's mapping after apply
- **AND** `legacy_field` SHALL remain present in the live cluster's mapping (it is not deleted)
