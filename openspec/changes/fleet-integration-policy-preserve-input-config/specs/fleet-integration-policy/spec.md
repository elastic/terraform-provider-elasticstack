## ADDED Requirements

### Requirement: Update — preserve input and stream `config` written outside Terraform (REQ-031)

The resource SHALL NOT remove input-level or stream-level `config` values from a package policy when it updates the policy. This data is written by Kibana or other Kibana features and is not managed by the resource.

#### Scenario: APM API keys survive a Terraform update

- **GIVEN** an APM integration policy whose input `config["apm-server"].value` contains `rum.source_mapping.elasticsearch.api_key` and `agent.config.elasticsearch.api_key` written by Kibana
- **WHEN** a Terraform update changes an input var
- **THEN** the updated policy SHALL contain the changed var
- **AND** the input `config` SHALL still contain both API keys

### Requirement: Update — snapshot `config` before the update (REQ-032)

Before submitting the update, the resource SHALL read the package policy in the typed (non-simplified) format and record the `config` of every input, identified by `type` and `policy_template`, and of every stream, identified additionally by `data_stream.dataset`. When that read fails, the resource SHALL return the error and SHALL NOT submit the update.

#### Scenario: No config means no extra write

- **GIVEN** an integration policy whose inputs and streams have no `config`
- **WHEN** Terraform updates the policy
- **THEN** exactly one update request SHALL be sent

### Requirement: Update — restore `config` removed by the update (REQ-033)

After the update succeeds, the resource SHALL read the policy again in the typed format and deep-merge each recorded `config` into the matching input or stream. A key present before and absent or null now SHALL be restored; when both values are objects the merge SHALL recurse; otherwise the current value SHALL be kept.

#### Scenario: Values written during the update are kept

- **GIVEN** an input `config` with keys `a` and `b`
- **WHEN** the update leaves `config` containing only `b` with a new value
- **THEN** the restored `config` SHALL contain `a` with its prior value and `b` with the new value

### Requirement: Update — restoring write (REQ-034)

When at least one value was restored, the resource SHALL submit one typed update containing only the `version` token from the preceding read and the full `inputs` array with compiled fields removed. When nothing was restored, it SHALL NOT submit a further update. Errors from the restoring read or update SHALL be returned.

#### Scenario: Config survives unchanged

- **GIVEN** an integration policy whose input `config` is not affected by the update
- **WHEN** Terraform updates the policy
- **THEN** no restoring update request SHALL be sent

#### Scenario: Restoring write carries only version and inputs

- **GIVEN** an update that removed input `config`
- **WHEN** the resource restores it
- **THEN** the restoring request body SHALL contain only `version` and `inputs`
- **AND** fields such as `name`, `description` and `policy_ids` SHALL be left unchanged on the server
