## Why

Practitioners who upload ML trained models to Elasticsearch (via Eland or out-of-band) cannot manage the deployment lifecycle (start, scale, stop) through Terraform today ([terraform-provider-elasticstack#725](https://github.com/elastic/terraform-provider-elasticstack/issues/725)). Users must manually call `POST _ml/trained_models/{model_id}/deployment/_start`, scale allocations via `_update`, and stop via `_stop`. This prevents infrastructure-as-code management of ML inference workloads.

Teams managing ML inference at scale (semantic search, NLP, embedding generation) need to control deployment parameters—allocation counts, thread counts, queue capacity, priority, and adaptive allocations—alongside the rest of their Elastic Stack infrastructure.

## What Changes

Add a new resource `elasticstack_elasticsearch_ml_trained_model_deployment` that manages the deployment lifecycle of an existing Elasticsearch ML trained model. The resource maps to:

- **Create** → `POST _ml/trained_models/{model_id}/deployment/_start`
- **Read** → `GET _ml/trained_models/{model_id}/_stats`
- **Update** → `POST _ml/trained_models/{deployment_id}/deployment/_update`
- **Delete** → `POST _ml/trained_models/{deployment_id}/deployment/_stop`

This follows the **state-transition resource** pattern already used by `elasticstack_elasticsearch_ml_job_state` and `elasticstack_elasticsearch_ml_datafeed_state`.

### Schema sketch

```hcl
resource "elasticstack_elasticsearch_ml_trained_model_deployment" "example" {
  id = <computed, string>  # <cluster_uuid>/<deployment_id>

  model_id      = <required, string>  # ForceNew; the model to deploy
  deployment_id = <optional+computed, string>  # ForceNew; defaults to model_id if omitted

  number_of_allocations = <optional, int>      # updatable; suppressed when adaptive_allocations.enabled=true
  threads_per_allocation = <optional, int>     # ForceNew
  priority               = <optional, string>  # ForceNew; "low" | "normal"
  queue_capacity         = <optional, int>     # ForceNew
  wait_for               = <optional, string>  # ForceNew; "starting" | "started" | "fully_allocated"
  api_timeout            = <optional, string>  # ForceNew; duration string for start API timeout

  adaptive_allocations {  # optional, updatable
    enabled                  = <required, bool>
    min_number_of_allocations = <optional, int>
    max_number_of_allocations = <optional, int>
  }

  timeouts {  # optional
    create = <optional, string>  # default: 5 minutes
    update = <optional, string>  # default: 5 minutes
  }

  # Computed
  state             = <computed, string>  # e.g. "started", "starting", "stopped"
  allocation_status = <computed, string>  # allocation status from stats
  stats_json        = <computed, string>  # raw JSON of deployment stats for extensibility
}
```

### Drift handling

When `adaptive_allocations.enabled = true`, the server controls `number_of_allocations`. The resource SHALL suppress plan diffs on `number_of_allocations` in this case to avoid perpetual drift. When `adaptive_allocations.enabled = false`, `number_of_allocations` is authoritative.

### Acceptance tests

- Requires a pre-existing trained model and an ML-enabled cluster.
- Tests MUST be skipped if no ML nodes are available or no suitable model exists.
- Separate test steps SHOULD cover: create with verification of computed attributes, re-plan with no diff, update `number_of_allocations`, update `adaptive_allocations`, import by composite ID, and delete (stop).

## Capabilities

### New Capabilities

- `elasticsearch-ml-trained-model-deployment`: New resource for starting, scaling, and stopping a trained model deployment; includes adaptive-allocations management and drift suppression for server-managed allocation counts.

### Modified Capabilities

- _(none)_

## Impact

- **Specs**: Delta under `openspec/changes/elasticsearch-ml-trained-model-deployment/specs/elasticsearch-ml-trained-model-deployment/spec.md` until synced into canonical spec.
- **Implementation** (future): new package `internal/elasticsearch/ml/trainedmodeldeployment/`, new client wrappers in `internal/clients/elasticsearch/`, registration in `provider/plugin_framework.go`.
