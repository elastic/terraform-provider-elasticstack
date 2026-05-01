## Why

All 39 `elasticstack_elasticsearch_ingest_processor_*` data sources are implemented using the Terraform Plugin SDK. They follow identical patterns: define a schema, read config into a model, marshal JSON, hash it for the computed `id`. This repetition makes the package unnecessarily large and hard to maintain. Before migrating the bulk of processors, we need to establish a shared generic base that eliminates per-processor `Read` duplication and validates the pattern with representative processors spanning different complexity levels.

## What Changes

- **Add** `internal/elasticsearch/ingest/processor_datasource_base.go` with:
  - `ProcessorModel` interface (`TypeName()`, `MarshalBody()`, `SetID()`, `SetJSON()`)
  - Generic `processorDataSource[T ProcessorModel]` struct owning `Metadata`, `Read`, `Configure`
  - `marshalAndHash()` helper wrapping body as `{"<name>": body}`, indent-marshaling, and hashing
- **Add** `internal/elasticsearch/ingest/processor_common.go` with:
  - `CommonProcessorModel` (PF struct for `description`, `if`, `ignore_failure`, `on_failure`, `tag`)
  - `CommonProcessorSchemaAttributes()` factory
  - `toCommonProcessorBody()` helper for `MarshalBody()` reuse
- **Add** `internal/elasticsearch/ingest/processor_models.go` with local inner structs for the 4 representatives
- **Migrate** 4 representative processor data sources to Plugin Framework:
  - `drop` — simplest possible (only common fields, no specific attributes)
  - `append` — medium complexity (specific fields + common, `ListAttribute`, defaults)
  - `script` — validators (`ExactlyOneOf` on `script_id` vs `source`), JSON blob `params`
  - `foreach` — JSON blob field `processor` parsing
- **Register** the 4 new constructors in `provider/plugin_framework.go`
- **Remove** the 4 old SDK registrations from `provider/provider.go`

## Capabilities

### New Capabilities

_None — this is an internal implementation refactor._

### Modified Capabilities

_None — no user-facing requirements change for the 4 representative processors._

## Impact

- `internal/elasticsearch/ingest/`: +3 shared base files, +4 new PF data source files
- `provider/plugin_framework.go`: +4 data source registrations
- `provider/provider.go`: -4 data source registrations
- `internal/models/ingest.go`: unaffected in this PR (structs remain in place)
