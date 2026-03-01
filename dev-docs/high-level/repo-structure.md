# Repo structure

For a more complete overview, see “Project Structure” in [`CONTRIBUTING.md`](../../CONTRIBUTING.md#project-structure).

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

