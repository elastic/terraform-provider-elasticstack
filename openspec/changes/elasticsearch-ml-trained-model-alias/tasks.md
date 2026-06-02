## 1. OpenSpec delta spec

- [x] 1.1 Create `openspec/changes/elasticsearch-ml-trained-model-alias/specs/elasticsearch-ml-trained-model-alias/spec.md` with full schema and requirements for the resource

## 2. Client wrappers

- [x] 2.1 Create `internal/clients/elasticsearch/ml_trained_model_alias.go` with:
  - `PutMLTrainedModelAlias(ctx, client, modelID, alias string, reassign bool) diag.Diagnostics`
  - `GetMLTrainedModelAlias(ctx, client, alias string) (modelID string, found bool, diags diag.Diagnostics)` — calls `GetTrainedModels` with alias as model_id, extracts model_id from response, returns not-found on empty result or 404
  - `DeleteMLTrainedModelAlias(ctx, client, alias string) diag.Diagnostics` — first resolves the alias via `GetTrainedModels` to obtain the current model_id, then calls DELETE; treats GET 404/empty and DELETE 404 as idempotent success

## 3. TF model and schema

- [x] 3.1 Create `internal/elasticsearch/ml/trainedmodelalias/models.go` with `TFModel` struct:
  - `ID types.String` (tfsdk:"id") — composite `<cluster_uuid>/<model_alias>`
  - `ElasticsearchConnection types.List` (tfsdk:"elasticsearch_connection")
  - `ModelAlias types.String` (tfsdk:"model_alias") — ForceNew, resource identity
  - `ModelID types.String` (tfsdk:"model_id") — mutable referent
  - `Reassign types.Bool` (tfsdk:"reassign")
  - Value-receiver methods: `GetID()`, `GetResourceID()` (returns ModelAlias), `GetElasticsearchConnection()`
- [x] 3.2 Create `internal/elasticsearch/ml/trainedmodelalias/schema.go` returning schema with:
  - `id` (computed, UseStateForUnknown)
  - `model_alias` (required, RequiresReplace)
  - `model_id` (required)
  - `reassign` (optional, bool, default true)

## 4. CRUD callbacks

- [x] 4.1 Create `internal/elasticsearch/ml/trainedmodelalias/create.go` — calls `PutMLTrainedModelAlias` with modelID from plan, alias from plan, reassign from plan; sets composite `id` via `client.ID(ctx, alias).String()`
- [x] 4.2 Create `internal/elasticsearch/ml/trainedmodelalias/read.go` — calls `GetMLTrainedModelAlias`; on found, populates `model_id` from response; on not-found, returns `(state, false, nil)`
- [x] 4.3 Create `internal/elasticsearch/ml/trainedmodelalias/update.go` — calls `PutMLTrainedModelAlias` with planned model_id, alias from resource identity, reassign from plan
- [x] 4.4 Create `internal/elasticsearch/ml/trainedmodelalias/delete.go` — calls `DeleteMLTrainedModelAlias` with alias from resource identity; the client wrapper resolves the current model_id and handles not-found

## 5. Resource registration

- [x] 5.1 Create `internal/elasticsearch/ml/trainedmodelalias/resource.go` — registers via `entitycore.NewElasticsearchResource[TFModel]` with create/read/update/delete; implements ImportState (passthrough on `id`)
- [x] 5.2 Create `internal/elasticsearch/ml/trainedmodelalias/descriptions.go` (and optional `descriptions/resource-description.md`)
- [x] 5.3 Register `trainedmodelalias.NewTrainedModelAliasResource` in `provider/plugin_framework.go` resources list and imports

## 6. Acceptance tests

- [x] 6.1 Create `internal/elasticsearch/ml/trainedmodelalias/acc_test.go` with:
  - TestAccResourceMLTrainedModelAlias_basic — create alias, verify id set, re-plan (no diff), import by composite id, delete
  - TestAccResourceMLTrainedModelAlias_reassign — create alias pointing to model A, update model_id to model B, verify model_id updated in state (default reassign=true)
  - TestAccResourceMLTrainedModelAlias_collisionWithReassignDisabled — attempt to create alias that already exists with reassign=false; expect error
  - TestAccResourceMLTrainedModelAlias_updateReassignFlag — change reassign from true to false on an existing resource

## 7. Verify

- [x] 7.1 `make build` passes
- [x] 7.2 Acceptance tests pass against a live Elasticsearch cluster (requires `TF_ACC=1` and ES 8.0+)
- [x] 7.3 `OPENSPEC_TELEMETRY=0 ./node_modules/.bin/openspec validate elasticsearch-ml-trained-model-alias --type change` passes
