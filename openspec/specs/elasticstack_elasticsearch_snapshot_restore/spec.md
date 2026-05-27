# elasticstack_elasticsearch_snapshot_restore Specification

## Purpose
TBD - created by archiving change elasticsearch-snapshot-actions. Update Purpose after archive.
## Requirements
### Requirement: Snapshot restore action (REQ-RESTORE)

The provider SHALL expose a Terraform provider-defined action named `elasticstack_elasticsearch_snapshot_restore` that invokes `POST /_snapshot/{repository}/{snapshot}/_restore` on an Elasticsearch cluster. **Requires Terraform 1.14+** (provider-defined actions are a Terraform Core 1.14+ feature).

The action MUST implement `action.Action` and `action.ActionWithConfigure` from `terraform-plugin-framework v1.19.0`. The provider MUST implement `provider.ProviderWithActions` and register this action.

**REQ-RESTORE-001**: The action SHALL invoke `POST /_snapshot/{repository}/{snapshot}/_restore` using the configured Elasticsearch client.

**REQ-RESTORE-002**: When `wait_for_completion` is `true`, the action SHALL pass `wait_for_completion=true` as a query parameter and block until the restore completes or the `invoke` timeout elapses.

**REQ-RESTORE-003**: When `wait_for_completion` is `false`, the action SHALL pass `wait_for_completion=false` and return immediately after the API accepts the request.

**REQ-RESTORE-004**: When the `invoke` timeout elapses before the restore completes, the action SHALL return a diagnostic error.

**REQ-RESTORE-005**: When the Elasticsearch API returns an error (including attempting to restore over existing indices without a rename), the action SHALL surface the full error message as a Terraform diagnostic error.

**REQ-RESTORE-006**: The `index_settings` attribute SHALL be accepted as a JSON-encoded string and passed to the API body as the `index_settings` field.

**REQ-RESTORE-007**: The `timeouts` block SHALL be implemented using the `action/timeouts` package from `terraform-plugin-framework-timeouts v0.7.0`.

**REQ-RESTORE-008**: Each action invocation SHALL be independent; re-applying the same configuration will re-invoke the restore.

**Schema:**

| Attribute | Type | Required | Description |
|---|---|---|---|
| `repository` | `string` | Required | Name of the snapshot repository |
| `snapshot` | `string` | Required | Name of the snapshot to restore |
| `indices` | `list(string)` | Optional | Index patterns to restore (all if omitted) |
| `include_global_state` | `bool` | Optional | Restore cluster state. Default: `false` |
| `ignore_unavailable` | `bool` | Optional | Ignore unavailable indices. Default: `false` |
| `include_aliases` | `bool` | Optional | Restore index aliases. Default: `true` |
| `partial` | `bool` | Optional | Allow partial restore. Default: `false` |
| `feature_states` | `list(string)` | Optional | Feature states to restore |
| `rename_pattern` | `string` | Optional | Regex pattern for renaming restored indices |
| `rename_replacement` | `string` | Optional | Replacement string for `rename_pattern` |
| `ignore_index_settings` | `list(string)` | Optional | Index settings to ignore on restore |
| `index_settings` | `string` (JSON) | Optional | JSON-encoded index settings overrides |
| `wait_for_completion` | `bool` | Optional | Wait for completion. Default: `true` |
| `timeouts.invoke` | `string` | Optional | Timeout duration, e.g. `"30m"`. Default: `"20m"` |
| `elasticsearch_connection` | block | Optional | Connection override |

#### Scenario: Synchronous restore completes successfully

- **GIVEN** a snapshot repository and a valid snapshot exist in Elasticsearch
- **AND** `wait_for_completion = true`
- **WHEN** the action is invoked
- **THEN** the action SHALL block until the restore completes and return with no diagnostic errors

#### Scenario: Restore fails due to existing indices without rename

- **GIVEN** the target indices already exist and no `rename_pattern` is configured
- **WHEN** the action is invoked
- **THEN** the action SHALL return a diagnostic error containing the Elasticsearch error message

#### Scenario: Asynchronous restore returns immediately

- **GIVEN** `wait_for_completion = false`
- **WHEN** the action is invoked
- **THEN** the action SHALL return as soon as the API accepts the request, without waiting for completion

#### Scenario: Restore with rename avoids conflict

- **GIVEN** the original indices exist in the cluster
- **AND** `rename_pattern` and `rename_replacement` are configured to produce non-conflicting index names
- **WHEN** the action is invoked
- **THEN** the action SHALL complete without error and the renamed indices SHALL exist

#### Scenario: Invoke timeout exceeded

- **GIVEN** `wait_for_completion = true` and `timeouts.invoke` is set to a value shorter than the restore duration
- **WHEN** the action is invoked
- **THEN** the action SHALL return a diagnostic error indicating the timeout was exceeded

