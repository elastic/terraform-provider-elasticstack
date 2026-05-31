# elasticsearch-content-connector-sync-job Specification

## Purpose
TBD - created by archiving change elasticsearch-content-connector. Update Purpose after archive.
## Requirements
### Requirement: Sync job create action (REQ-SYNC-001)

The provider SHALL expose a Terraform provider-defined action named `elasticstack_elasticsearch_connector_sync_job_create` that invokes `POST /_connector/_sync_job` to create a sync job for an existing connector. The action complements the schedule-driven sync flow for on-demand syncs. **Requires Terraform 1.14+** (provider-defined actions are a Terraform Core 1.14+ feature).

The action MUST implement `action.Action` and `action.ActionWithConfigure` from `terraform-plugin-framework v1.19.0`. The provider MUST implement `provider.ProviderWithActions` and register this action.

**REQ-SYNC-001-A**: The action SHALL invoke `POST /_connector/_sync_job` using the configured Elasticsearch client with a body containing `id` (the `connector_id`), `job_type`, and `trigger_method`.

**REQ-SYNC-001-B**: When `wait_for_completion = true`, the action SHALL poll `GET /_connector/_sync_job/{sync_job_id}` every 5 seconds until the job reaches a terminal status (`completed`, `cancelled`, `error`, `suspended`) OR the `timeouts.invoke` timeout elapses.

**REQ-SYNC-001-C**: When `wait_for_completion = false` (default), the action SHALL return immediately after the API accepts the create request.

**REQ-SYNC-001-D**: When the `invoke` timeout elapses before the sync job completes, the action SHALL return a diagnostic error naming the sync job ID and its last observed status.

**REQ-SYNC-001-E**: When the sync job reaches terminal status `error`, the action SHALL return a diagnostic error including the job's `error` field.

**REQ-SYNC-001-F**: When the sync job reaches terminal status `cancelled` or `suspended`, the action SHALL return a diagnostic error indicating the non-success terminal status.

**REQ-SYNC-001-G**: When the Elasticsearch API returns a non-success response (e.g. connector not found, missing privileges), the action SHALL surface the full error message as a Terraform diagnostic error.

**REQ-SYNC-001-H**: The `timeouts` block SHALL be implemented using the `action/timeouts` package from `terraform-plugin-framework-timeouts v0.7.0`. The default `timeouts.invoke` value SHALL be `30m`.

**REQ-SYNC-001-I**: Each action invocation SHALL be independent. Re-invoking the action SHALL create a new sync job each time.

**REQ-SYNC-001-J**: The action SHALL NOT delete the sync job document after completion or failure. Sync job history is preserved for operator inspection.

**Schema:**

| Attribute | Type | Required | Description |
|---|---|---|---|
| `connector_id` | `string` | Required | The id of the connector to sync. |
| `job_type` | `string` | Optional | One of `"full"`, `"incremental"`, `"access_control"`. Default `"full"`. |
| `trigger_method` | `string` | Optional | One of `"on_demand"`, `"scheduled"`. Default `"on_demand"`. |
| `wait_for_completion` | `bool` | Optional | When `true`, blocks until the sync job reaches a terminal status. Default `false`. |
| `timeouts.invoke` | `string` | Optional | Timeout duration when waiting for completion, e.g. `"60m"`. Default `"30m"`. |
| `elasticsearch_connection` | block | Optional | Connection override. |

#### Scenario: Asynchronous create returns immediately

- **GIVEN** `wait_for_completion = false`
- **WHEN** the action is invoked
- **THEN** the action SHALL call `POST /_connector/_sync_job` and return as soon as the API accepts the request without polling

#### Scenario: Synchronous create waits for completion

- **GIVEN** `wait_for_completion = true` and a connector whose service is running
- **WHEN** the action is invoked
- **THEN** the action SHALL poll the sync job's status every 5 seconds
- **AND** SHALL return successfully once the sync job reaches terminal status `completed`

#### Scenario: Synchronous create surfaces error status

- **GIVEN** `wait_for_completion = true`
- **AND** the sync job reaches terminal status `error` with `error = "permission denied"`
- **WHEN** the action observes that status
- **THEN** the action SHALL return a diagnostic error including `"permission denied"`

#### Scenario: Invoke timeout exceeded

- **GIVEN** `wait_for_completion = true` and `timeouts.invoke = "5s"`
- **AND** the sync job has not reached terminal status when the timeout elapses
- **WHEN** the action observes the timeout
- **THEN** the action SHALL return a diagnostic error naming the sync job ID and its last observed status

#### Scenario: Connector not found

- **GIVEN** `connector_id` references a connector that does not exist
- **WHEN** the action is invoked
- **THEN** the API SHALL return an error
- **AND** the action SHALL surface it as a Terraform diagnostic error

#### Scenario: Default job_type and trigger_method

- **GIVEN** `connector_id` is set and `job_type` and `trigger_method` are omitted
- **WHEN** the action is invoked
- **THEN** the request body SHALL include `job_type = "full"` and `trigger_method = "on_demand"`

#### Scenario: Sync job history preserved

- **GIVEN** the action completes successfully
- **WHEN** the action returns
- **THEN** the sync job document SHALL remain in the internal index
- **AND** the action SHALL NOT call `DELETE /_connector/_sync_job/{id}`

### Requirement: Sync job action minimum versions (REQ-SYNC-002)

The action SHALL require:
- Terraform Core 1.14.0+ (for provider-defined actions).
- Elasticsearch 8.16.0+ for the action. Although the connector resource itself is available from Elasticsearch 8.12.0 (sync job APIs were introduced in that release), the `POST /_connector/_sync_job` request body validation on 8.12.x–8.15.x rejects the on-wire field name produced by the current typed Elasticsearch Go client; the API stabilized on the field shape the provider sends starting with 8.16.0.

Older versions of either SHALL produce a clear diagnostic.

#### Scenario: Older Terraform Core diagnostic

- **GIVEN** Terraform Core older than 1.14.0
- **WHEN** the action is invoked
- **THEN** Terraform itself SHALL reject the action (no provider-defined-actions support)

#### Scenario: Older Elasticsearch diagnostic

- **GIVEN** the configured Elasticsearch is older than 8.16.0
- **WHEN** the action is invoked
- **THEN** diagnostics SHALL include an error stating the action requires Elasticsearch 8.16.0 or later

