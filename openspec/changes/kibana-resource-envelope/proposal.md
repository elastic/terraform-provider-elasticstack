## Why

Kibana resources repeat the same 5–6 line CRUD prelude in every lifecycle method — decode model, resolve scoped client via `kibana_connection`, unwrap `GetKibanaOapiClient`, extract `space_id` — with no shared abstraction to own it. `NewElasticsearchResource` already eliminates this boilerplate for Elasticsearch-backed resources; Kibana resources deserve the same.

## What Changes

- Add `KibanaResourceModel` interface to `entitycore` with `GetID`, `GetResourceID`, `GetSpaceID`, and `GetKibanaConnection` methods.
- Add `KibanaResource[T]` generic struct to `entitycore` that owns Schema (with `kibana_connection` block injection), Configure, Metadata, Create, Read, Update, and Delete.
- Add `NewKibanaResource[T]` constructor with callback arguments for read, delete, create, and update operations.
- Add `PlaceholderKibanaWriteCallbacks[T]` for parity with `PlaceholderElasticsearchWriteCallbacks[T]`.
- Add unit tests for `KibanaResource[T]` envelope in `entitycore` (mirroring `resource_envelope_test.go`).
- Migrate `internal/kibana/streams` to use `NewKibanaResource` as the user-specified-ID POC.
- Migrate `internal/kibana/maintenance_window` to use `NewKibanaResource` as the API-assigned-UUID POC.
- Update `entitycore/doc.go` to document the Kibana resource envelope pattern.

## Capabilities

### New Capabilities

- `entitycore-kibana-resource-envelope`: Generic Kibana resource envelope — `KibanaResourceModel` interface, `KibanaResource[T]` type, `NewKibanaResource` constructor, callback function types, `PlaceholderKibanaWriteCallbacks`, and the composite-ID-or-fallback prelude shared by Read, Update, and Delete.

### Modified Capabilities

## Impact

- `internal/entitycore/` — new file `kibana_resource_envelope.go`; updated `doc.go`
- `internal/kibana/streams/` — `resource.go`, `create.go`, `read.go`, `update.go`, `delete.go` simplified; model gains `GetResourceID`, `GetSpaceID`, `GetKibanaConnection` methods
- `internal/kibana/maintenance_window/` — same files simplified; model gains the same interface methods
