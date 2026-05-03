## Why

The provider is incrementally migrating Elasticsearch API calls from the raw `esapi` (untyped) client to the `go-elasticsearch` Typed API (`elasticsearch.TypedClient`). The transform helpers in `internal/clients/elasticsearch/transform.go` are the next surface to migrate. Using the typed client eliminates hand-structured JSON request bodies, removes manual JSON marshaling/unmarshaling boilerplate, and enables compile-time API shape checking.

## What Changes

- Migrate `internal/clients/elasticsearch/transform.go` helper functions to use the typed client:
  - `PutTransform` → typed `Transform.PutTransform` API
  - `GetTransform` → typed `Transform.GetTransform` API
  - `GetTransformStats` → typed `Transform.GetTransformStats` API
  - `UpdateTransform` → typed `Transform.UpdateTransform` API
  - `DeleteTransform` → typed `Transform.DeleteTransform` API
  - `startTransform` → typed `Transform.StartTransform` API
  - `stopTransform` → typed `Transform.StopTransform` API
- Replace custom model types (`models.Transform`, `models.TransformStats`, `models.PutTransformParams`, `models.UpdateTransformParams`, and related response wrappers) with their typed-client equivalents where possible.
- Update the `internal/elasticsearch/transform/transform.go` resource and test files to consume the migrated helpers.
- Ensure zero behavioral changes from the Terraform user perspective; this is a pure implementation refactor.

Files affected:
- `internal/clients/elasticsearch/transform.go`
- `internal/models/transform.go`
- `internal/elasticsearch/transform/transform.go`
- `internal/elasticsearch/transform/transform_test.go`

## Capabilities

### New Capabilities
_(none — this is an internal refactoring with no new user-visible capabilities)_

### Modified Capabilities
- `elasticsearch-transform`: Migrate internal Transform API client usage from raw `esapi` to the typed API. The Terraform resource schema and user-visible behavior remain identical, but the implementation contract (helper signatures, model structs, and response parsing paths) changes, warranting an updated spec.

## Impact

- **Code**: Transform helpers (`transform.go`), transform resource, and transform tests.
- **APIs**: No Terraform resource or data source behavior changes.
- **Dependencies**: Relies on existing `go-elasticsearch/v8` Typed API support via `ElasticsearchScopedClient.GetESTypedClient()`.
- **Build / CI**: Compilation and acceptance tests are affected; no new dependencies introduced.
