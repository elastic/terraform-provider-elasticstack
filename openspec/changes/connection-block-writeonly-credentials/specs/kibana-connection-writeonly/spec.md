## ADDED Requirements

### Requirement: `password_wo` write-only attribute in `kibana_connection`

The per-resource `kibana_connection` block SHALL include a `password_wo` attribute that is `Optional`, `Sensitive`, and `WriteOnly`. It SHALL NOT be stored in Terraform state after apply. It SHALL be available in config and plan context during CRUD operations. It SHALL conflict with the plain `password` attribute.

#### Scenario: `password_wo` accepted and not stored in state
- **WHEN** a resource is configured with `password_wo = "<secret>"` in its `kibana_connection` block
- **THEN** Terraform SHALL accept the configuration
- **AND** the applied state SHALL NOT contain the `password_wo` value

#### Scenario: `password_wo` and `password` cannot be set together
- **WHEN** both `password` and `password_wo` are set in the same `kibana_connection` block
- **THEN** Terraform SHALL reject the configuration with a conflict error

### Requirement: `api_key_wo` write-only attribute in `kibana_connection`

The per-resource `kibana_connection` block SHALL include an `api_key_wo` attribute that is `Optional`, `Sensitive`, and `WriteOnly`. It SHALL NOT be stored in Terraform state after apply. It SHALL conflict with the plain `api_key` attribute.

#### Scenario: `api_key_wo` accepted and not stored in state
- **WHEN** a resource is configured with `api_key_wo = "<key>"` in its `kibana_connection` block
- **THEN** Terraform SHALL accept the configuration
- **AND** the applied state SHALL NOT contain the `api_key_wo` value

#### Scenario: `api_key_wo` and `api_key` cannot be set together
- **WHEN** both `api_key` and `api_key_wo` are set in the same `kibana_connection` block
- **THEN** Terraform SHALL reject the configuration with a conflict error

### Requirement: `bearer_token_wo` write-only attribute in `kibana_connection`

The per-resource `kibana_connection` block SHALL include a `bearer_token_wo` attribute that is `Optional`, `Sensitive`, and `WriteOnly`. It SHALL NOT be stored in Terraform state after apply. It SHALL conflict with the plain `bearer_token` attribute.

#### Scenario: `bearer_token_wo` accepted and not stored in state
- **WHEN** a resource is configured with `bearer_token_wo = "<token>"` in its `kibana_connection` block
- **THEN** Terraform SHALL accept the configuration
- **AND** the applied state SHALL NOT contain the `bearer_token_wo` value

#### Scenario: `bearer_token_wo` and `bearer_token` cannot be set together
- **WHEN** both `bearer_token` and `bearer_token_wo` are set in the same `kibana_connection` block
- **THEN** Terraform SHALL reject the configuration with a conflict error

### Requirement: `PreferWriteOnlyAttribute` validators on plain credential companions

Each plain credential attribute in `kibana_connection` that has a `_wo` sibling (`password`, `api_key`, `bearer_token`) SHALL include a `PreferWriteOnlyAttribute` validator pointing at the corresponding `_wo` companion.

#### Scenario: Plan-time warning when plain attribute is used instead of `_wo`
- **WHEN** `api_key` is set in the `kibana_connection` block and `api_key_wo` is available
- **THEN** Terraform SHALL emit a plan-time warning recommending the use of `api_key_wo` instead

### Requirement: `_wo` value wins over plain when Kibana client is built

When building the Kibana client from a per-resource `kibana_connection` block, the factory SHALL use the `_wo` value for each credential field when it is non-empty, falling back to the plain value otherwise.

#### Scenario: `api_key_wo` used over `api_key` when both non-empty
- **WHEN** both `api_key` and `api_key_wo` are non-empty (despite the conflict warning)
- **THEN** the Kibana client SHALL be built using the `api_key_wo` value

#### Scenario: Plain `api_key` used when `api_key_wo` is absent
- **WHEN** only `api_key` is set in the `kibana_connection` block
- **THEN** the Kibana client SHALL be built using the plain `api_key` value

### Requirement: Drift detection for `_wo` attributes via `writeonlyhash`

The Kibana resource envelope SHALL implement `ModifyPlan` that uses `internal/utils/writeonlyhash` to detect silent in-config changes to each `_wo` credential attribute (`password_wo`, `api_key_wo`, `bearer_token_wo`). On detecting a changed `_wo` value (hash mismatch), the envelope SHALL emit a warning diagnostic naming the attribute path only (no value) and mark the resource for update.

#### Scenario: Unchanged `_wo` value produces no drift signal
- **WHEN** the same `api_key_wo` value is set in config between two applies
- **THEN** `ModifyPlan` SHALL NOT emit a warning
- **AND** no update SHALL be scheduled based on the `_wo` value alone

#### Scenario: Changed `_wo` value triggers warning and update
- **WHEN** `api_key_wo` in config is changed between two applies
- **THEN** `ModifyPlan` SHALL emit a warning diagnostic naming the `api_key_wo` attribute path
- **AND** an update SHALL be scheduled

#### Scenario: Hash stored after successful Create
- **WHEN** a resource with `api_key_wo` is successfully created
- **THEN** the bcrypt hash of the `api_key_wo` value SHALL be stored in resource private state

#### Scenario: Hash cleared after Delete
- **WHEN** a resource with `api_key_wo` is destroyed
- **THEN** all `_wo`-related private-state hash entries SHALL be cleared

### Requirement: Backward compatibility

The new per-resource `kibana_connection` block function SHALL include all attributes present in the existing provider-schema variant. No existing attribute may be removed or renamed.

#### Scenario: Existing config with plain attributes continues to work
- **WHEN** an existing resource configuration uses `api_key` in `kibana_connection` and `api_key_wo` is not set
- **THEN** Terraform SHALL accept the configuration and authenticate using the plain value

### Requirement: No `_wo_version` attributes

Write-only version companion attributes (e.g., `api_key_wo_version`) SHALL NOT be added to the block. Drift detection SHALL rely exclusively on the `writeonlyhash` private-state mechanism.

#### Scenario: No `_wo_version` attribute in schema
- **WHEN** the resource schema is inspected
- **THEN** no attribute ending in `_wo_version` SHALL appear in the `kibana_connection` block
