# Design: Migrate simple PF resources to entitycore envelope

## Overview

Four Plugin Framework resources currently embed `*entitycore.ResourceBase` directly. Each has standard CRUD that fits the envelope callback contract completely. This change migrates all four to `*entitycore.ElasticsearchResource[Data]`.

## Resources

### 1. index_alias (`internal/elasticsearch/index/alias/`)
- Current: PF ResourceBase with custom Create/Update/Delete/Read
- Create: `UpdateIndexAlias` (PUT-like)
- Update: `DeleteIndexAlias` + `UpdateIndexAlias`
- Read: `GetIndexAlias` or `GetIndex` to read aliases
- Delete: `DeleteIndexAlias`
- Model type: `models.go` has the current model — add getters and convert Read/Delete to callbacks.

### 2. data_stream_lifecycle (`internal/elasticsearch/index/datastreamlifecycle/`)
- Current: PF ResourceBase with custom Create/Update/Delete/Read
- Simple PUT lifecycle policy body
- Read: `GetDataStreamLifecycle`
- Delete: `DeleteDataStreamLifecycle` (or reset to default)
- Model type exists — add getters.

### 3. enrich_policy (`internal/elasticsearch/enrich/`)
- Current: PF ResourceBase with custom Schema, Create (PUT + optional Execute), Update, Delete
- Has `execute` bool that triggers `ExecuteEnrichPolicy` after PUT.
- Create callback: PUT policy, then if `execute=true`, call Execute API, then set ID.
- Update callback: same as create (recreate on change, since most fields are ForceNew).
- Read: `GetEnrichPolicy`
- Delete: `DeleteEnrichPolicy`
- Model: enrich already has a model. Add `GetID()`, etc.
- **Note**: `ImportState` currently sets `execute=true` after passthrough. Preserve this on concrete type.

### 4. inference_endpoint (`internal/elasticsearch/inference/inferenceendpoint/`)
- Current: PF ResourceBase with custom Create/Update/Delete/Read
- PUT endpoint creation
- Read: `GetInferenceEndpoint`
- Delete: `DeleteInferenceEndpoint`
- Model exists — add getters. All attributes map cleanly.

## Common Pattern

For each resource:
1. Add `GetID()`, `GetResourceID()`, `GetElasticsearchConnection()` to model.
2. Convert existing `Read` method body to package-level `readXxx(ctx, client, id, state) (T, bool, diag.Diagnostics)`.
3. Convert existing `Delete` method body to package-level `deleteXxx(ctx, client, id, state) diag.Diagnostics`.
4. Convert existing `Create` method body to package-level `createXxx(ctx, client, id, state) (T, diag.Diagnostics)`.
5. Convert existing `Update` method body to package-level `updateXxx` callback.
6. In resource constructor, call `entitycore.NewElasticsearchResource[Data]` with callbacks.
7. Strip `elasticsearch_connection` block from schema factory (envelope injects).
8. Keep `ImportState` as a method on concrete type (envelope does not provide it).

## Schema Block Injection

The envelope's `Schema` method copies blocks and injects `elasticsearch_connection`. Current resources that define it manually in `GetResourceSchema()` must remove it.

## Testing

- No Terraform interface changes.
- Acceptance tests for each resource validate correctness.
