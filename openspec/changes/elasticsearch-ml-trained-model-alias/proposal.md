## Why

Trained model aliases provide a stable logical name that can be referenced in inference configurations, ingest processors, and aggregations. When a model is replaced, reassigning the alias avoids updating every reference to the model. There is currently no Terraform resource for managing this lifecycle, so alias assignment must be done via direct API calls.

## What Changes

Add a new resource `elasticstack_elasticsearch_ml_trained_model_alias` that wraps the Elasticsearch ML trained model alias API:

- `PUT _ml/trained_models/{model_id}/model_aliases/{model_alias}` — Create and Update
- `DELETE _ml/trained_models/{model_id}/model_aliases/{model_alias}` — Delete
- `GET _ml/trained_models/{model_alias}` — Read (resolves alias to model_id)

## Capabilities

### New Capabilities

- `elasticsearch-ml-trained-model-alias`: Full CRUD resource for ML trained model aliases. The alias name (`model_alias`) is the stable identifier; the referenced model (`model_id`) can be updated in-place using the PUT API's `reassign` query parameter. Out-of-band alias deletion is handled by re-creating on next apply. Out-of-band reassignment shows as a model_id diff on next plan.

## Impact

- New package `internal/elasticsearch/ml/trainedmodelalias/` (resource.go, schema.go, models.go, read.go, create.go, update.go, delete.go, acc_test.go, descriptions.go)
- New client wrappers in `internal/clients/elasticsearch/ml_trained_model_alias.go`
- Provider registration in `provider/plugin_framework.go`
- New spec at `openspec/specs/elasticsearch-ml-trained-model-alias/spec.md`
