# elasticstack_elasticsearch_snapshot_create Specification

## Purpose
TBD - created by archiving change elasticsearch-snapshot-actions. Update Purpose after archive.
## Requirements
### Requirement: On-demand snapshot create action (REQ-CREATE)

The provider SHALL expose a Terraform provider-defined action named `elasticstack_elasticsearch_snapshot_create` that invokes `POST /_snapshot/{repository}/{snapshot}` to create an Elasticsearch snapshot on demand. This action complements the SLM-managed (`elasticstack_elasticsearch_snapshot_lifecycle`) scheduled approach for DR workflows. **Requires Terraform 1.14+** (provider-defined actions are a Terraform Core 1.14+ feature).

The action MUST implement `action.Action` and `action.ActionWithConfigure` from `terraform-plugin-framework v1.19.0`. The provider MUST implement `provider.ProviderWithActions` and register this action.

**REQ-CREATE-001**: The action SHALL invoke `POST /_snapshot/{repository}/{snapshot}` using the configured Elasticsearch client.

**REQ-CREATE-002**: When `wait_for_completion` is `true`, the action SHALL pass `wait_for_completion=true` as a query parameter and block until snapshot creation completes or the `invoke` timeout elapses.

**REQ-CREATE-003**: When `wait_for_completion` is `false`, the action SHALL pass `wait_for_completion=false` and return immediately after the API accepts the request.

**REQ-CREATE-004**: When the `invoke` timeout elapses before the snapshot completes, the action SHALL return a diagnostic error.

**REQ-CREATE-005**: When the Elasticsearch API returns an error (e.g., duplicate snapshot name, repository not registered), the action SHALL surface the full error message as a Terraform diagnostic error.

**REQ-CREATE-006**: The `metadata` attribute SHALL be accepted as a JSON-encoded string and passed to the API body as the `metadata` field.

**REQ-CREATE-007**: The `timeouts` block SHALL be implemented using the `action/timeouts` package from `terraform-plugin-framework-timeouts v0.7.0`.

**REQ-CREATE-008**: Each action invocation SHALL be independent; re-applying the same configuration will attempt to create the snapshot again (the API will error if a snapshot with the same name already exists).

**Schema:**

| Attribute | Type | Required | Description |
|---|---|---|---|
| `repository` | `string` | Required | Name of the snapshot repository |
| `snapshot` | `string` | Required | Name to assign to the snapshot |
| `indices` | `list(string)` | Optional | Index patterns to include (all if omitted) |
| `include_global_state` | `bool` | Optional | Include cluster state. Default: `false` |
| `ignore_unavailable` | `bool` | Optional | Ignore unavailable indices. Default: `false` |
| `partial` | `bool` | Optional | Allow partial snapshot. Default: `false` |
| `feature_states` | `list(string)` | Optional | Feature states to include |
| `expand_wildcards` | `string` | Optional | Wildcard expansion: `"open"`, `"closed"`, `"hidden"`, `"none"`, `"all"`. Default: `"open"` |
| `metadata` | `string` (JSON) | Optional | JSON-encoded metadata to attach to the snapshot |
| `wait_for_completion` | `bool` | Optional | Wait for completion. Default: `true` |
| `timeouts.invoke` | `string` | Optional | Timeout duration, e.g. `"60m"`. Default: `"20m"` |
| `elasticsearch_connection` | block | Optional | Connection override |

#### Scenario: Synchronous snapshot creation completes successfully

- **GIVEN** a registered snapshot repository exists in Elasticsearch
- **AND** `wait_for_completion = true`
- **WHEN** the action is invoked with a unique snapshot name
- **THEN** the action SHALL block until the snapshot is created and return with no diagnostic errors

#### Scenario: Snapshot creation fails due to duplicate name

- **GIVEN** a snapshot with the same name already exists in the repository
- **WHEN** the action is invoked
- **THEN** the action SHALL return a diagnostic error containing the Elasticsearch error message

#### Scenario: Asynchronous creation returns immediately

- **GIVEN** `wait_for_completion = false`
- **WHEN** the action is invoked
- **THEN** the action SHALL return as soon as the API accepts the request, without waiting for completion

#### Scenario: Snapshot created with metadata

- **GIVEN** `metadata = jsonencode({ created_by = "terraform", env = "prod" })`
- **WHEN** the action is invoked
- **THEN** the snapshot SHALL be created with the provided metadata attached and no diagnostic errors SHALL occur

#### Scenario: Invoke timeout exceeded

- **GIVEN** `wait_for_completion = true` and `timeouts.invoke` is set to a value shorter than the snapshot creation duration
- **WHEN** the action is invoked
- **THEN** the action SHALL return a diagnostic error indicating the timeout was exceeded

