## Why

Agent Builder data sources repeat the same read pipeline—config decode, scoped client resolution, version enforcement, composite ID splitting, space ID fallback, and state setting—with only the final API call and model mapping varying (#2511). This duplication is repeated to a lesser degree across all Kibana and Elasticsearch data sources. Existing `DataSourceBase` eliminates `Configure` and `Metadata` duplication, but every data source still rewrites its entire `Read()` method from scratch.

## What Changes

- Introduce generic `entitycore.NewKibanaDataSource[T]()` and `entitycore.NewElasticsearchDataSource[T]()` constructors that own the full Terraform contract (`Configure`, `Metadata`, `Schema`, `Read`) via an envelope pattern.
- The base injects the scoped connection block (`kibana_connection` / `elasticsearch_connection`) into the schema automatically.
- The concrete package provides only:
  1. **Schema** (without connection block)
  2. **Model** (without connection field)
  3. **A pure read function** — entity-specific API call and state mapping
- Migrate the three Agent Builder data sources identified in #2511 to the new pattern as reference implementations.
- No changes to resource patterns or SDK-based data sources.

## Capabilities

### New Capabilities
- `entitycore-datasource-envelope`: Generic envelope constructor for Plugin Framework data sources that eliminates Read() orchestration boilerplate by owning config decode, scoped client resolution, and state persistence. Concrete implementations supply only schema and a pure read function.

### Modified Capabilities
<!-- No spec-level requirement changes. Agent builder data sources, kibana spaces, and fleet enrollment tokens change only implementation architecture, not observable behavior. -->

## Impact

- **Package**: `internal/entitycore` — new generic constructors and envelope types.
- **Packages**: `internal/kibana/agentbuilderagent`, `internal/kibana/agentbuildertool`, `internal/kibana/agentbuilderworkflow` — migrated to envelope pattern.
- **No breaking changes** to Terraform schemas, state shapes, or provider API.
- Existing struct-based data sources remain fully functional.
