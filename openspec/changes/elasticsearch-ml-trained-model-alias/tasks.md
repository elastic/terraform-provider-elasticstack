## 1. OpenSpec delta spec

- [ ] 1.1 Create `openspec/changes/elasticsearch-ml-trained-model-alias/specs/elasticsearch-ml-trained-model-alias/spec.md` with full schema and requirements for the resource

## 2. Client wrappers

- [ ] 2.1 Create `internal/clients/elasticsearch/ml_trained_model_alias.go` with:
  - `PutMLTrainedModelAlias(ctx, client, modelID, alias string, reassign bool) diag.Diagnostics`
  - `GetMLTrainedModelAlias(ctx, client, alias string) (modelID string, found bool, diags diag.Diagnostics)` — calls `GetTrainedModels` with alias as model_id, extracts model_id from response, returns not-found on empty result or 404
  - `DeleteMLTrainedModelAlias(ctx, client, modelID, alias string) diag.Diagnostics` — treats 404 as idempotent success

## 3. TF model and schema

- [ ] 3.1 Create `internal/elasticsearch/ml/trainedmodelalias/models.go` with `TFModel` struct:
  - `ID types.String` (tfsdk:"id") — composite `<cluster_uuid>/<model_alias>`
  - `ElasticsearchConnection types.List` (tfsdk:"elasticsearch_connection")
  - `ModelAlias types.String` (tfsdk:"model_alias") — ForceNew, resource identity
  - `ModelID types.String` (tfsdk:"model_id") — mutable referent
  - `Reassign types.Bool` (tfsdk:"reassign")
  - Value-receiver methods: `GetID()`, `GetResourceID()` (returns ModelAlias), `GetElasticsearchConnection()`
- [ ] 3.2 Create `internal/elasticsearch/ml/trainedmodelalias/schema.go` returning schema with:
  - `id` (computed, UseStateForUnknown)
  - `model_alias` (required, RequiresReplace)
  - `model_id` (required)
  - `reassign` (optional, bool, default false)

## 4. CRUD callbacks

- [ ] 4.1 Create `internal/elasticsearch/ml/trainedmodelalias/create.go` — calls `PutMLTrainedModelAlias` with modelID from plan, alias from plan, reassign from plan; sets composite `id` via `client.ID(ctx, alias).String()`
- [ ] 4.2 Create `internal/elasticsearch/ml/trainedmodelalias/read.go` — calls `GetMLTrainedModelAlias`; on found, populates `model_id` from response; on not-found, returns `(state, false, nil)`
- [ ] 4.3 Create `internal/elasticsearch/ml/trainedmodelalias/update.go` — calls `PutMLTrainedModelAlias` with planned model_id, alias from resource identity, reassign from plan
- [ ] 4.4 Create `internal/elasticsearch/ml/trainedmodelalias/delete.go` — calls `DeleteMLTrainedModelAlias` with model_id from prior state, alias from resource identity

## 5. Resource registration

- [ ] 5.1 Create `internal/elasticsearch/ml/trainedmodelalias/resource.go` — registers via `entitycore.NewElasticsearchResource[TFModel]` with create/read/update/delete; implements ImportState (passthrough on `id`)
- [ ] 5.2 Create `internal/elasticsearch/ml/trainedmodelalias/descriptions.go` (and optional `descriptions/resource-description.md`)
- [ ] 5.3 Register `trainedmodelalias.NewTrainedModelAliasResource` in `provider/plugin_framework.go` resources list and imports

## 6. Acceptance tests

- [ ] 6.1 Create `internal/elasticsearch/ml/trainedmodelalias/acc_test.go` with:
  - TestAccResourceMLTrainedModelAlias_basic — create alias, verify id set, re-plan (no diff), import by composite id, delete
  - TestAccResourceMLTrainedModelAlias_reassign — create alias pointing to model A, update model_id to model B with reassign=true, verify model_id updated in state
  - TestAccResourceMLTrainedModelAlias_collisionWithoutReassign — attempt to create alias that already exists without reassign=true; expect error
  - TestAccResourceMLTrainedModelAlias_updateReassignFlag — change reassign from false to true on an existing resource

## 7. Verify

- [ ] 7.1 `make build` passes
- [ ] 7.2 Acceptance tests pass against a live Elasticsearch cluster (requires `TF_ACC=1` and ES 8.0+)
- [ ] 7.3 `OPENSPEC_TELEMETRY=0 ./node_modules/.bin/openspec validate elasticsearch-ml-trained-model-alias --type change` passes
