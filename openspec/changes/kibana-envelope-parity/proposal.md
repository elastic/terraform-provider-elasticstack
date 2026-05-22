## Why

The Kibana resource envelope (`entitycore.KibanaResource[T]`) is structurally behind the Elasticsearch envelope: it lacks enforced read-after-write, a unified write callback type, decoded Terraform config in write requests, and a post-read hook. This means Kibana resources either carry their own ad-hoc read-after-write logic inside callbacks (e.g., `maintenance_window`) or silently persist potentially stale state after writes.

## What Changes

- **BREAKING** Replace `KibanaCreateFunc[T]` and `KibanaUpdateFunc[T]` with a unified `KibanaWriteFunc[T]`, matching `WriteFunc[T]` in the Elasticsearch envelope. Create and Update are distinguished by `req.Prior == nil`.
- **BREAKING** Replace positional parameters in `NewKibanaResource` with `KibanaResourceOptions[T]`, matching the ES envelope's `ElasticsearchResourceOptions[T]`.
- **BREAKING** `PlaceholderKibanaWriteCallbacks[T]()` changes from returning `(KibanaCreateFunc[T], KibanaUpdateFunc[T])` to returning a single `KibanaWriteFunc[T]`.
- Add `KibanaWriteRequest[T]` struct carrying `Plan`, `Prior *T`, `Config`, `WriteID`, and `SpaceID`.
- Add `KibanaWriteResult[T]` struct carrying the written model used for read-after-write identity resolution.
- Add `KibanaPostReadFunc[T]` and a `PostRead` field in `KibanaResourceOptions[T]` (optional).
- Enforce read-after-write: after every Create and Update, the envelope calls `readFunc` and persists the read result into state. If not found after write, a "Resource not found after write" error is returned.
- Pass decoded Terraform `Config T` to write callbacks (required for write-only attributes).
- Invoke `PostRead` hook after every successful read (including read-after-write) when configured.
- Migrate all 6 concrete Kibana resources to the new callback contract. `maintenance_window` is simplified — its manual internal read-after-write is removed as the envelope absorbs the responsibility.

## Capabilities

### New Capabilities

None — this change modifies an existing capability.

### Modified Capabilities

- `entitycore-kibana-resource-envelope`: New requirements for enforced read-after-write, unified `KibanaWriteFunc[T]`, `KibanaWriteRequest[T]` / `KibanaWriteResult[T]` types, `Config T` in write requests, `PostRead` hook, and `KibanaResourceOptions[T]` constructor.

## Impact

- `internal/entitycore/kibana_resource_envelope.go` — primary change
- `internal/entitycore/kibana_resource_envelope_test.go` — new tests for all new behaviors
- `internal/fleet/proxy/` — callback signatures updated
- `internal/kibana/streams/` — callback signatures updated
- `internal/kibana/maintenance_window/` — callback signatures updated; manual read-after-write removed from `create.go`
- `internal/kibana/spaces/` — callback signatures updated
- `internal/kibana/security_role/` — callback signatures updated
- `internal/fleet/agentdownloadsource/` — `PlaceholderKibanaWriteCallbacks` call-site updated (no callback body changes; resource still overrides Create/Update directly)
