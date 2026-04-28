# Delta Spec: `elasticstack_elasticsearch_index` Template Mapping Drift

Base spec: `openspec/specs/elasticsearch-index/spec.md`
Resource implementation: `internal/elasticsearch/index/index`
Last requirement in base spec: REQ-029
This delta modifies: REQ-015-REQ-024

---

This delta defines the target behavior introduced by change `fix-index-template-mapping-drift`. It describes the requirements the implementation will satisfy once the change is applied; it is not intended to claim that the current implementation already meets these behaviors.

## MODIFIED Requirements

### Requirement: Update flow (REQ-015–REQ-018)

On update, the resource SHALL only call the relevant update APIs when the corresponding values have changed. Alias changes SHALL be applied by deleting aliases removed from config (via Delete Alias API) and upserting all aliases present in plan (via Put Alias API). Dynamic setting changes SHALL be applied by calling the Put Settings API with the diff, setting removed dynamic settings to `null` in the request. Mapping changes SHALL be applied by calling the Put Mapping API only when the user-owned mapping intent has semantically changed. Template-injected mapping content that appears in the Elasticsearch Get Index API response SHALL NOT by itself cause a mapping update, replacement, provider inconsistent-result error, or non-empty follow-up plan. All update APIs SHALL target the persisted concrete index identity from state / `id`, not the configured `name`. After all updates, the resource SHALL perform a read to refresh state while preserving any configured `name` already stored in state.

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

### Requirement: Read (REQ-019–REQ-021)

On read, the resource SHALL parse `id` to extract the concrete index name, call the Get Index API with `flat_settings=true`, and if the index is not found (HTTP 404 or missing from response), SHALL remove the resource from state without error. When the index is found, the resource SHALL populate `concrete_name`, all aliases, `mappings`, `settings_raw`, and all individual setting attributes from the API response. For `mappings`, read SHALL preserve the user's prior mapping intent when the API response is a semantically equal superset caused by mappings injected by a matching index template. When state already contains a configured `name`, read SHALL preserve that configured value and SHALL NOT overwrite it with the concrete index name. When state does not contain `name`, read SHALL backfill `name` from the concrete index name.

#### Scenario: Index not found

- **WHEN** the Get Index API returns 404
- **THEN** the resource SHALL remove itself from state without error

#### Scenario: Date math name remains stable during read

- **WHEN** state already contains a configured date math expression in `name` and read refreshes the managed concrete index
- **THEN** `name` SHALL remain unchanged and `concrete_name` SHALL reflect the concrete index being managed

#### Scenario: Template-only mappings stay non-drifting

- **GIVEN** an index resource has no configured `mappings`
- **AND** a matching index template injects mappings into the created index
- **WHEN** read refreshes the index and Terraform plans the unchanged configuration
- **THEN** Terraform SHALL produce an empty plan for the index resource

#### Scenario: User-owned mappings tolerate template-injected extras

- **GIVEN** an index resource has configured `mappings`
- **AND** a matching index template injects additional mapping `properties`, `dynamic_templates`, or other top-level mapping keys
- **WHEN** read refreshes the index after create or during a later plan
- **THEN** Terraform SHALL NOT report a provider inconsistent-result error
- **AND** Terraform SHALL produce an empty plan for the unchanged configuration

### Requirement: Mappings plan modifier and semantic equality (REQ-022–REQ-024)

The `mappings` attribute SHALL use shared mapping comparison semantics for both semantic equality and replacement decisions. The comparison SHALL preserve existing mapped fields not present in config when those fields are user-owned and Elasticsearch would retain them after a field removal request. When a user-owned field is removed from config `mappings.properties`, the provider SHALL add a warning diagnostic and retain the field in the planned value or otherwise treat the retained field as semantically equal state. When a user-owned field's `type` changes between state and config, the provider SHALL require replacement. When `mappings.properties` is removed entirely from config while user-owned properties are present in state, the provider SHALL require replacement.

For mapping content injected by a matching index template, including additional `properties`, `dynamic_templates`, `_meta`, `runtime`, or other top-level mapping keys absent from user configuration, the resource SHALL treat the API value as a non-drifting superset of the user-owned mapping intent. The resource SHALL NOT require `lifecycle.ignore_changes = [mappings]` to avoid drift caused only by those template-injected mappings.

For `semantic_text` fields, Elasticsearch automatically enriches the stored mapping with a `model_settings` object (containing inference model configuration such as `dimensions`, `element_type`, `service`, `similarity`, and `task_type`) after index creation. When the field type in state and config is `semantic_text` and `model_settings` is present in state but absent from the config, the provider SHALL treat the enriched mapping as semantically equal to the configured mapping so that the plan matches the value Elasticsearch will return. When `model_settings` is explicitly specified in config, the config value SHALL be used as-is and SHALL NOT be overwritten by the state value.

#### Scenario: Field removed from config

- GIVEN state `mappings` contains user-owned field `foo` and config `mappings` does not
- WHEN plan runs
- THEN the plan SHALL retain `foo` in the planned `mappings` or treat the retained state value as semantically equal
- AND the provider SHALL add a warning diagnostic

#### Scenario: Field type changed

- GIVEN state `mappings` has user-owned field `foo` with `type: keyword` and config has `type: text`
- WHEN plan runs
- THEN the provider SHALL mark the resource for replacement

#### Scenario: semantic_text field without explicit model_settings in config

- GIVEN state `mappings` contains a `semantic_text` field with server-enriched `model_settings`
- AND the config for that field does not specify `model_settings`
- WHEN plan runs
- THEN the provider SHALL treat the server-enriched `model_settings` as semantically equal to the configured field

#### Scenario: semantic_text field with explicit model_settings in config

- GIVEN state `mappings` contains a `semantic_text` field with `model_settings`
- AND the config for that field also specifies `model_settings`
- WHEN plan runs
- THEN the provider SHALL use the config `model_settings` value and SHALL NOT overwrite it with the state value

#### Scenario: Template-injected dynamic templates are non-drift

- **GIVEN** a matching index template injects `dynamic_templates`
- **AND** the index resource configuration does not own those `dynamic_templates`
- **WHEN** Terraform compares refreshed mappings with prior user intent
- **THEN** the template-injected `dynamic_templates` SHALL be treated as non-drift
