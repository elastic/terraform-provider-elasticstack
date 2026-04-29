# Repo structure

For a more complete overview, see "Project Structure" in [`contributing.md`](./contributing.md#project-structure).

## Data source patterns

There are two supported patterns for Plugin Framework data sources. Choose the one that fits the entity's Read complexity.

### Envelope generics (simple data sources)

Use [entitycore.NewKibanaDataSource] or [entitycore.NewElasticsearchDataSource] when the data source follows the standard read pipeline without conditional early returns or complex state manipulation:

1. Decode config
2. Resolve scoped client
3. Call API
4. Map response to model
5. Set state

The envelope owns steps 1, 2, and 5. The concrete package provides only a schema factory (without connection blocks), a model embedding [entitycore.KibanaConnectionField] or [entitycore.ElasticsearchConnectionField], and a pure read function.

Good fits: single-entity lookups (e.g., Agent Builder workflow, Agent Builder tool, spaces list).

### Struct-based embedding (complex data sources)

Embed [*entitycore.DataSourceBase] and implement [datasource.DataSource] directly when the Read method needs:

- Conditional early returns or branching API calls
- Custom state manipulation beyond simple model mapping
- Multiple API calls with interdependent logic (e.g., tool dependency graphs)
- Special diagnostic handling that doesn't fit the uniform pipeline

Good fits: Agent Builder agent (tool dependency graph, version-gated workflow embedding), multi-step lookups.

## Key directories

- `internal/`: provider implementation (Go)
  - `internal/elasticsearch/`: Elasticsearch-specific resources and logic
  - `internal/kibana/`: Kibana-specific resources and logic
  - `internal/fleet/`: Fleet-specific resources and logic
- `provider/`: provider wiring and configuration
- `generated/`: generated API clients (notably `generated/kbapi`)
- `docs/`: generated provider docs
- `templates/`: docs templates
- `examples/`: examples used by docs generation and as references
- `scripts/`: dev/CI helper scripts

