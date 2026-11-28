## ADDED Requirements

### Requirement: Update — preserve input and stream `config` written outside Terraform (REQ-031)

The resource SHALL NOT remove input-level or stream-level `config` values from a package policy when it updates the policy. This data is written by Kibana or other Kibana features and is not managed by the resource.

#### Scenario: APM API keys survive a Terraform update

- **GIVEN** an APM integration policy whose input `config["apm-server"].value` contains `rum.source_mapping.elasticsearch.api_key` and `agent.config.elasticsearch.api_key` written by Kibana
- **WHEN** a Terraform update changes an input var
- **THEN** the updated policy SHALL contain the changed var
- **AND** the input `config` SHALL still contain both API keys

### Requirement: Update — snapshot `config` before the update (REQ-032)

Before submitting the update, the resource SHALL read the package policy in the typed format and record its `revision` and the `config` of every input (identified by `type` and `policy_template`) and stream (identified additionally by `data_stream.dataset`). An identity shared by more than one input or stream, with or without `config`, SHALL NOT be recorded. When that read fails or does not find the policy, the resource SHALL return an error and SHALL NOT submit the update.

#### Scenario: No config means no extra write

- **GIVEN** an integration policy whose inputs and streams have no `config`
- **WHEN** Terraform updates the policy
- **THEN** exactly one update request SHALL be sent

#### Scenario: Snapshot read fails

- **GIVEN** the typed read before the update fails or does not find the policy
- **WHEN** Terraform updates the policy
- **THEN** the resource SHALL return an error
- **AND** no update request SHALL be sent

### Requirement: Update — restore `config` removed by the update (REQ-033)

After the update succeeds, the resource SHALL read the policy again in the typed format and deep-merge each recorded `config` into the matching input or stream, only when that identity is unique in the updated policy. A key present before and absent now SHALL be restored, including a key whose prior value is null; a key null now SHALL be restored when its prior value is not null; when both values are objects the merge SHALL recurse; otherwise the current value SHALL be kept.

#### Scenario: Ambiguous identities are never restored

- **GIVEN** two inputs with the same `type` and `policy_template`, only one of them with `config`
- **WHEN** Terraform updates the policy
- **THEN** no `config` SHALL be recorded or restored for that identity

#### Scenario: Values written during the update are kept

- **GIVEN** an input `config` with keys `a` and `b`
- **WHEN** the update leaves `config` containing only `b` with a new value
- **THEN** the restored `config` SHALL contain `a` with its prior value and `b` with the new value

#### Scenario: Stream config is restored

- **GIVEN** a stream with `data_stream.dataset` `d` whose `config` contains key `k`
- **WHEN** the update leaves that stream without `config`
- **THEN** the restored stream `config` SHALL contain `k` with its prior value

#### Scenario: A key left null by the update is restored

- **GIVEN** an input `config` whose key `a` has a non-null value
- **WHEN** the update leaves `a` null
- **THEN** the restored `config` SHALL contain `a` with its prior value

#### Scenario: Nested keys with null values are restored

- **GIVEN** an input `config` whose object `parent` contains `kept` and `removed`, where `removed` is null
- **WHEN** the update leaves `parent` containing only `kept`
- **THEN** the restored `parent` SHALL contain `kept` and `removed` with a null value

### Requirement: Update — restoring write (REQ-034)

When at least one value was restored, the resource SHALL submit one typed update containing only the `version` token, `policy_id` and `policy_ids` from the preceding read and the full `inputs` array with compiled fields removed. When nothing was restored, or the request would carry the same `inputs` as the policy already holds, it SHALL NOT submit a further update. Errors from the restoring read or update SHALL be returned, and a policy that is not found by the restoring read SHALL be an error.

#### Scenario: Config survives unchanged

- **GIVEN** an integration policy whose input `config` is not affected by the update
- **WHEN** Terraform updates the policy
- **THEN** no restoring update request SHALL be sent

#### Scenario: Restored null that the typed request cannot carry

- **GIVEN** recorded `config` whose entry had a null `value`, which the typed request omits
- **WHEN** the update removed only that `value`
- **THEN** no restoring update request SHALL be sent

#### Scenario: Policy deleted before the restoring read

- **GIVEN** recorded `config` and a policy that is not found by the restoring read
- **WHEN** the resource restores the `config`
- **THEN** it SHALL return an error and SHALL NOT submit an update

#### Scenario: Restoring write carries only version, agent policy assignment and inputs

- **GIVEN** an update that removed input `config`
- **WHEN** the resource restores it
- **THEN** the restoring request body SHALL contain only `version`, `policy_id`, `policy_ids` and `inputs`
- **AND** fields not in the request, such as `name` and `description`, SHALL be left unchanged on the server
- **AND** `policy_id` and `policy_ids` SHALL keep the values from the preceding read

### Requirement: Update — never restore from a stale snapshot (REQ-035)

When a restoring write is needed, the resource SHALL compare the `revision` of the restoring read with the `revision` recorded by the snapshot. Kibana increments the revision by one on every update and ignores the `version` token on a simplified update, so any revision other than the recorded one plus one means another client wrote the policy after the snapshot. In that case the resource SHALL return an error and SHALL NOT submit the restoring update.

#### Scenario: Pack removed between the snapshot and the update

- **GIVEN** recorded `config` containing Osquery packs `pack-1` and `pack-2`
- **AND** another client removes `pack-2` after the snapshot and before the simplified update
- **WHEN** the resource restores the `config`
- **THEN** it SHALL return an error
- **AND** it SHALL NOT submit a restoring update, so `pack-2` is not brought back

### Requirement: Update — report `config` that could not be restored (REQ-036)

When the restoring read, the stale-snapshot check or the restoring update fails after the simplified update was applied, the resource SHALL return, in addition to the error of the failing step, an error stating that `config` removed by the update was not written back and that re-running apply does not restore it.

#### Scenario: Restoring read fails after the update

- **GIVEN** recorded `config` and a simplified update that succeeded
- **WHEN** the restoring read fails
- **THEN** the resource SHALL return the read error
- **AND** an error stating that the removed `config` was not written back and that re-running apply does not restore it

#### Scenario: Restoring update fails after the update

- **GIVEN** recorded `config` that the simplified update removed
- **WHEN** the restoring update fails
- **THEN** the resource SHALL return the restoring update error
- **AND** an error stating that the removed `config` was not written back and that re-running apply does not restore it

#### Scenario: Stale snapshot detected after the update

- **GIVEN** recorded `config` that the simplified update removed and a restoring read whose `revision` is not the recorded one plus one
- **WHEN** the resource restores the `config`
- **THEN** it SHALL return the stale-snapshot error
- **AND** an error stating that the removed `config` was not written back and that re-running apply does not restore it
