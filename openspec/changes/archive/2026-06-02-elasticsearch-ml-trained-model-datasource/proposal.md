## Why

Terraform configurations for semantic search clusters cannot reference trained models that were uploaded via Eland or provisioned out-of-band. There is currently no way to read a trained model into the Terraform dependency graph, which means downstream resources — ML trained model deployments, inference ingest processors, and related aliases — cannot reference model metadata in a type-safe, plan-visible way.

This is part of issue #725 (ML APIs for semantic search). This proposal covers only the **read-only data source**; the write resource for model upload is deferred.

## What Changes

- Add a new **data source** `elasticstack_elasticsearch_ml_trained_model` that wraps `GET _ml/trained_models/{model_id}` and exposes safe computed metadata fields.
- The data source reads one model by its `model_id` (or alias) and surfaces fields needed by downstream Terraform resources.
- **Out of scope for this proposal**: editing `openspec/specs/` directly; that happens when the change is synced or archived.

### Schema sketch

```hcl
data "elasticstack_elasticsearch_ml_trained_model" "example" {
  model_id = "my-model"  # required

  # Computed attributes
  id                     = <computed, string>   # "<cluster_uuid>/<model_id>"
  description            = <computed, string>
  model_type             = <computed, string>   # "lang_ident" | "ner" | "text_classification" | "pytorch" | …
  model_size_bytes       = <computed, int>
  fully_defined          = <computed, bool>
  tags                   = <computed, set(string)>
  create_time            = <computed, string>   # ISO-8601 timestamp
  created_by             = <computed, string>
  version                = <computed, string>   # Elasticsearch version
  platform_architecture  = <computed, string>
  license_level          = <computed, string>
  input_json             = <computed, string>   # JSON of model input field names
  inference_config_json  = <computed, string>   # JSON of default inference configuration
  metadata_json          = <computed, string>   # JSON of model metadata map
  default_field_map      = <computed, map(string)>

  # Resource-level Elasticsearch connection override (injected by entitycore)
  elasticsearch_connection { … }
}
```

Definition fields (`compressed_definition`, `definition`) are explicitly excluded — they are large, write-only, and not needed for a read data source. No secrets are surfaced.

### Version requirements

The ML trained models API is available from Elasticsearch 8.0+. No specific minimum version gate is expected beyond what the existing provider configuration enforces. This should be confirmed during implementation.

### Acceptance tests

- Acceptance tests must pre-seed or skip based on an existing trained model in the cluster (e.g. a built-in model, or one pre-loaded by the test environment).
- Test scenarios: basic read by `model_id`, graceful 404 handling, and (if supported) read by model alias.

## Capabilities

### New Capabilities

- `elasticsearch-ml-trained-model`: Read-only data source for Elasticsearch ML trained model configuration — `GET _ml/trained_models/{model_id}`, computed metadata fields, graceful 404 handling, and acceptance-test expectations.

### Modified Capabilities

- _(none)_

## Impact

- **Specs**: Delta under `openspec/changes/elasticsearch-ml-trained-model-datasource/specs/elasticsearch-ml-trained-model/spec.md` until merged into canonical spec.
- **Implementation** (future): new package `internal/elasticsearch/ml/trainedmodel/` (data_source.go, read.go, models.go, schema.go), optional client helper in `internal/clients/elasticsearch/`, registration in `provider/plugin_framework.go`.
