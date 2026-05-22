## Why

`fleet/proxy` and `fleet/agentdownloadsource` each carry a hand-written version check (`assertVersionSupported`) duplicated across every CRUD method, and own all CRUD boilerplate manually rather than delegating to the `entitycore.KibanaResource` envelope that every comparable Kibana-backed resource already uses. Migrating them eliminates the per-operation version check duplication and aligns them with the established `NewKibanaResource` pattern.

## What Changes

- **`internal/fleet/proxy`**: `Resource` changes from embedding `*entitycore.ResourceBase` to `*entitycore.KibanaResource[proxyModel]`. All four CRUD receiver methods become package-level envelope callbacks. `proxyModel` gains `KibanaResourceModel` interface methods and `WithVersionRequirements`. `version.go` is deleted. `kibana_connection` block is removed from the schema factory (envelope injects it). `ImportState` is retained as a concrete method on `Resource`.
- **`internal/fleet/agentdownloadsource`**: `Resource` changes from embedding `*entitycore.ResourceBase` to `*entitycore.KibanaResource[model]`. `Read` and `Delete` become envelope callbacks; `Create` and `Update` remain concrete override methods (they require reading operational space from state rather than plan, and `Create` has a read-back-after-write pattern). `model` gains `KibanaResourceModel` interface methods and `WithVersionRequirements`. `GetSpaceID()` returns `"default"` when `space_ids` is null or empty (safe: `BuildSpaceAwarePath` treats `""` and `"default"` identically). `version.go` is deleted. `kibana_connection` block removed from schema factory. `SpaceImporter` updated to populate both `source_id` and `id` on import.
- Version enforcement on Delete is intentionally dropped in both resources (the check is not meaningful in a delete path and no existing resource enforces it).
- `entitycore_contract_test.go` in `agentdownloadsource` updated: embed assertion updated from `ResourceBase` to `KibanaResource`, and the import test gains an assertion for the `id` field.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `fleet-agent-download-source`: The minimum version guard is narrowed from "plan or apply" to Create, Read, and Update only. Delete no longer enforces the version check (the `KibanaResource` envelope does not invoke version requirements on Delete, and enforcing it on destroy is not meaningful).

## Impact

- `internal/fleet/proxy`: `resource.go`, `models.go`, `schema.go`, `create.go`, `read.go`, `update.go`, `delete.go`, `version.go` (deleted)
- `internal/fleet/agentdownloadsource`: `resource.go`, `models.go`, `schema.go`, `create.go`, `read.go`, `update.go`, `delete.go`, `version.go` (deleted), `entitycore_contract_test.go`
- No provider API surface changes. No acceptance test changes required.
