## MODIFIED Requirements

### Requirement: Mappings plan modifier (REQ-022–REQ-024)

The `mappings` attribute SHALL use a custom plan modifier that preserves existing mapped fields not present in config, because Elasticsearch ignores field removal requests. When a field is removed from config `mappings.properties`, the plan modifier SHALL add a warning diagnostic and retain the field in the planned value. When a field's `type` changes between state and config, the plan modifier SHALL require replacement. When `mappings.properties` is removed entirely from config while present in state, the plan modifier SHALL require replacement.

For `semantic_text` fields, Elasticsearch automatically enriches the stored mapping with a `model_settings` object (containing inference model configuration such as `dimensions`, `element_type`, `service`, `similarity`, and `task_type`) after index creation. When the field type in state and config is `semantic_text` and `model_settings` is present in state but absent from the config, the plan modifier SHALL copy `model_settings` from state into the planned value so that the plan matches the value Elasticsearch will return. When `model_settings` is explicitly specified in config, the config value SHALL be used as-is and SHALL NOT be overwritten by the state value.

#### Scenario: Field removed from config

- GIVEN state `mappings` contains field `foo` and config `mappings` does not
- WHEN plan runs
- THEN the plan SHALL retain `foo` in the planned `mappings` and SHALL add a warning diagnostic

#### Scenario: Field type changed

- GIVEN state `mappings` has field `foo` with `type: keyword` and config has `type: text`
- WHEN plan runs
- THEN the plan modifier SHALL mark the resource for replacement

#### Scenario: semantic_text field without explicit model_settings in config

- GIVEN state `mappings` contains a `semantic_text` field with server-enriched `model_settings`
- AND the config for that field does not specify `model_settings`
- WHEN plan runs
- THEN the plan modifier SHALL copy `model_settings` from state into the planned value

#### Scenario: semantic_text field with explicit model_settings in config

- GIVEN state `mappings` contains a `semantic_text` field with `model_settings`
- AND the config for that field also specifies `model_settings`
- WHEN plan runs
- THEN the plan modifier SHALL use the config `model_settings` value and SHALL NOT overwrite it with the state value
