# `elasticstack_elasticsearch_ml_trained_model_deployment` — Schema and Functional Requirements

Resource implementation: `internal/elasticsearch/ml/trainedmodeldeployment`

## Purpose

Manage the deployment lifecycle (start, scale, stop) of an existing Elasticsearch ML trained model. This resource does not upload or create the underlying trained model; it only manages the deployment state. On Terraform destroy the resource stops (undeploys) the model deployment.

## Schema

```hcl
resource "elasticstack_elasticsearch_ml_trained_model_deployment" "example" {
  id = <computed, string>  # internal identifier: <cluster_uuid>/<deployment_id>

  model_id      = <required, string>  # ForceNew; the trained model to deploy
  deployment_id = <optional+computed, string>  # ForceNew; custom deployment identifier; defaults to model_id when omitted

  number_of_allocations  = <optional, int>     # updatable; number of model allocations; suppressed when adaptive_allocations.enabled=true
  threads_per_allocation = <optional, int>     # ForceNew; threads used per allocation during inference
  priority               = <optional, string>  # ForceNew; "low" | "normal"
  queue_capacity         = <optional, int>     # ForceNew; max queued inference requests
  wait_for               = <optional, string>  # "starting" | "started" | "fully_allocated"; allocation state to wait for; default: "fully_allocated"
  api_timeout            = <optional, string>  # Go duration string; server-side start timeout
  force_stop             = <optional, bool>    # default: false; pass force=true to the Stop API on destroy

  adaptive_allocations {  # optional, updatable
    enabled                   = <required, bool>
    min_number_of_allocations = <optional, int>
    max_number_of_allocations = <optional, int>
  }

  timeouts {  # optional
    create = <optional, string>  # default: 5 minutes; total Terraform wait for start including polling
    update = <optional, string>  # default: 5 minutes
  }

  # Computed
  state             = <computed, string>  # deployment state: "starting" | "started" | "stopping" | "stopped"
  allocation_status = <computed, string>  # allocation status from deployment stats
  stats_json        = <computed, string>  # raw JSON of TrainedModelStats for extensibility

  elasticsearch_connection {  # optional, deprecated
    endpoints                = <optional, list(string)>
    username                 = <optional, string>
    password                 = <optional, string>
    api_key                  = <optional, string>
    bearer_token             = <optional, string>
    es_client_authentication = <optional, string>
    insecure                 = <optional, bool>
    headers                  = <optional, map(string)>
    ca_file                  = <optional, string>
    ca_data                  = <optional, string>
    cert_file                = <optional, string>
    key_file                 = <optional, string>
    cert_data                = <optional, string>
    key_data                 = <optional, string>
  }
}
```

## Requirements

### Requirement: API — Start, Update, Stop, and Stats (REQ-001–REQ-004)

To deploy a trained model, the resource SHALL call `POST _ml/trained_models/{model_id}/deployment/_start` ([docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/start-trained-model-deployment.html)). To update a deployed model, the resource SHALL call `POST _ml/trained_models/{deployment_id}/deployment/_update` ([docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/update-trained-model-deployment.html)). To read current deployment state, the resource SHALL call `GET _ml/trained_models/{model_id}/_stats` ([docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/get-trained-models-stats.html)). To stop a deployment, the resource SHALL call `POST _ml/trained_models/{deployment_id}/deployment/_stop` ([docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/stop-trained-model-deployment.html)). When any of these APIs returns a non-success response, the resource SHALL surface the error in Terraform diagnostics.

#### Scenario: Start API error surfaced

- GIVEN the Start Deployment API returns a non-success response
- WHEN create runs
- THEN Terraform diagnostics SHALL include the API error

#### Scenario: Update API error surfaced

- GIVEN the Update Deployment API returns a non-success response
- WHEN update runs
- THEN Terraform diagnostics SHALL include the API error

#### Scenario: Stop API error surfaced

- GIVEN the Stop Deployment API returns a non-success response (other than 404)
- WHEN delete runs
- THEN Terraform diagnostics SHALL include the API error

#### Scenario: Stop 404 treated as success

- GIVEN the Stop Deployment API returns HTTP 404
- WHEN delete runs
- THEN the resource SHALL be removed from state without error (idempotent delete)

#### Scenario: Force stop on destroy

- GIVEN a deployment with `force_stop = true`
- WHEN delete runs
- THEN the Stop Deployment API SHALL be called with `force=true`

### Requirement: Identity (REQ-005)

The resource SHALL expose a computed `id` attribute in the format `<cluster_uuid>/<deployment_id>`. The `deployment_id` SHALL default to the value of `model_id` when not provided. Both `id` and `deployment_id` SHALL use `UseStateForUnknown` to preserve values across plans.

#### Scenario: ID set after create

- GIVEN a successful start operation
- WHEN create completes
- THEN `id` in state SHALL be `<cluster_uuid>/<deployment_id>`

#### Scenario: deployment_id defaults to model_id

- GIVEN `deployment_id` is not configured by the practitioner
- WHEN create runs
- THEN `deployment_id` in state SHALL equal `model_id`

### Requirement: Import (REQ-006)

The resource SHALL support import via `ImportStatePassthroughID` on the `id` attribute using the composite format `<cluster_uuid>/<deployment_id>`. After import, subsequent reads SHALL populate all computed attributes.

#### Scenario: Import by composite ID

- GIVEN a running deployment with a known composite ID
- WHEN imported via `terraform import ... <cluster_uuid>/<deployment_id>`
- THEN state SHALL reflect the current deployment state, deployment_id, state, and allocation_status

### Requirement: ForceNew attributes (REQ-007)

The attributes `model_id`, `deployment_id`, `threads_per_allocation`, `priority`, and `queue_capacity` SHALL be marked `RequiresReplace`. Any change to these attributes SHALL trigger destroy-then-create (stop the existing deployment and start a new one).

#### Scenario: threads_per_allocation change triggers replace

- GIVEN a deployed model with `threads_per_allocation = 1`
- WHEN configuration changes `threads_per_allocation` to `2`
- THEN Terraform plan SHALL show a replace operation (destroy + create)

### Requirement: Wait-for polling (REQ-008)

When `wait_for` is configured, the resource SHALL poll `GET _ml/trained_models/{model_id}/_stats` after start until the deployment's `allocation_status.state` matches the configured `wait_for` value, or until the duration specified by `timeouts.create` elapses. On timeout, the resource SHALL return a diagnostic error without rolling back the deployment.

#### Scenario: wait_for = "started" — deployment reaches started

- GIVEN `wait_for = "started"` configured
- WHEN the deployment reaches `allocation_status.state = "started"` before `timeouts.create` elapses
- THEN create SHALL succeed and state SHALL reflect `state = "started"`

#### Scenario: wait_for timeout

- GIVEN `wait_for = "fully_allocated"` and `timeouts.create = "30s"` (short)
- WHEN the timeout elapses before the deployment reaches "fully_allocated"
- THEN the resource SHALL return a diagnostic error

### Requirement: Mutable attributes — Update (REQ-009)

Changes to `number_of_allocations` and `adaptive_allocations` SHALL call `POST _ml/trained_models/{deployment_id}/deployment/_update` without destroying and recreating the deployment.

#### Scenario: Update number_of_allocations

- GIVEN a running deployment with `number_of_allocations = 1`
- WHEN configuration changes `number_of_allocations` to `2`
- THEN Terraform plan SHALL show an in-place update (no replace)
- AND update SHALL call the Update Deployment API with `number_of_allocations = 2`

#### Scenario: Update adaptive_allocations enabled

- GIVEN a running deployment with `adaptive_allocations.enabled = false`
- WHEN configuration sets `adaptive_allocations.enabled = true`
- THEN Terraform plan SHALL show an in-place update
- AND update SHALL call the Update Deployment API with the new adaptive_allocations settings

### Requirement: Adaptive allocations (REQ-010)

The schema SHALL enforce mutual exclusivity between `number_of_allocations` and `adaptive_allocations` via a `ConflictsWith` validator. Practitioners SHALL configure exactly one of fixed allocations (`number_of_allocations`) or adaptive allocations (`adaptive_allocations`). When `adaptive_allocations.enabled = true`, the Elasticsearch server controls the effective `number_of_allocations`.

#### Scenario: Configuring both number_of_allocations and adaptive_allocations fails validation

- GIVEN a configuration with both `number_of_allocations = 1` and `adaptive_allocations.enabled = true`
 WHEN `terraform plan` runs
- THEN Terraform SHALL emit a validation error because the attributes are mutually exclusive

#### Scenario: Diff surfaces on number_of_allocations when fixed allocations used

- GIVEN `adaptive_allocations` is not configured and the API returns a different `number_of_allocations` than configured
- WHEN `terraform plan` runs after apply
- THEN the plan SHALL show a diff for `number_of_allocations`

#### Scenario: Switch from adaptive to fixed allocations triggers update

- GIVEN a running deployment with `adaptive_allocations.enabled = true`
- WHEN the configuration removes `adaptive_allocations` and sets `number_of_allocations = 2`
- THEN Terraform plan SHALL show an in-place update
- AND update SHALL call the Update Deployment API with `number_of_allocations = 2`

### Requirement: Computed attributes — state, allocation_status, stats_json (REQ-011)

After every create, update, and read operation, the resource SHALL populate:

- `state`: the deployment state string from `deployment_stats.state` in the stats response (e.g. `"started"`, `"starting"`, `"stopped"`).
- `allocation_status`: the allocation status state string from `deployment_stats.allocation_status.state`.
- `stats_json`: the full JSON serialisation of the `TrainedModelStats` entry for this deployment, for extensibility.

#### Scenario: Computed attributes populated after create

- GIVEN a successful deployment start
- WHEN create completes
- THEN `state`, `allocation_status`, and `stats_json` SHALL be non-empty in state

### Requirement: External stop detection (REQ-012)

If the deployment is stopped by an external actor (outside Terraform), the Read operation SHALL treat the deployment as not-found (no matching stats for `deployment_id`). The resource SHALL be removed from state with no error. The next `terraform plan` SHALL show a re-create and the next `terraform apply` SHALL call the Start API to restore the deployment.

#### Scenario: External stop detected on plan

- GIVEN a deployment stopped outside Terraform
- WHEN Read runs during refresh
- THEN the resource is removed from state
- AND the next plan shows a re-create

### Requirement: Minimum Elasticsearch version (REQ-013)

Trained model deployment APIs are GA from Elasticsearch 8.0. The provider SHALL support this resource on Elasticsearch 8.0 and later. No additional version gate beyond the provider's existing minimum ES version requirement is needed.

#### Scenario: Resource available on ES 8.0+

- GIVEN an Elasticsearch cluster running version 8.0 or later
- WHEN the resource is applied
- THEN the provider SHALL call the deployment APIs without emitting a version-gate error

### Requirement: Connection (REQ-015)

The resource SHALL support the standard (deprecated) `elasticsearch_connection` override block. When `elasticsearch_connection` is configured, the resource SHALL resolve the Elasticsearch client from the override rather than the provider-level configuration, matching the behavior of other ML state-transition resources (e.g. `elasticstack_elasticsearch_ml_job_state`).

#### Scenario: Resource-level connection override

- GIVEN a resource with `elasticsearch_connection` configured
- WHEN any API call is made
- THEN the resource SHALL use the overridden client settings

### Requirement: Acceptance tests (REQ-016)

The acceptance test suite SHALL:

1. Skip all tests if no ML nodes are available or no suitable trained model exists in the test environment.
2. Include a test that creates a deployment and verifies computed attributes (`state`, `deployment_id`, `id`).
3. Include a `PlanOnly` step after initial create to verify no diff.
4. Include a test that updates `number_of_allocations` and verifies the update is applied.
5. Include a test that enables `adaptive_allocations` and verifies `number_of_allocations` diff is suppressed on subsequent plan.
6. Include a test that imports the deployment by composite ID and verifies state.
7. Include a test that destroys the resource and verifies the deployment is stopped.
8. Include a test that attempts to deploy a non-existent model and verifies an error diagnostic is returned.

#### Scenario: Create and verify computed attributes

- GIVEN a valid `model_id` for an existing trained model
- WHEN the resource is applied
- THEN `state = "started"`, `deployment_id` is set, and `id = "<cluster_uuid>/<deployment_id>"`

#### Scenario: No diff on re-plan

- GIVEN a successfully applied deployment
- WHEN `terraform plan` runs again with no config changes
- THEN the plan SHALL be empty

#### Scenario: Import roundtrip

- GIVEN a running deployment with composite ID `<cluster_uuid>/<deployment_id>`
- WHEN imported and the plan run
- THEN state SHALL match and the plan SHALL be empty
