## Why

The provider has no way to manage [Elasticsearch content connectors](https://www.elastic.co/docs/reference/search-connectors) — the integrations that sync data from third-party sources (PostgreSQL, GitHub, SharePoint, S3, Salesforce, etc.) into Elasticsearch. Teams that adopt connectors today must script their lifecycle out-of-band with `curl` against the [connector APIs](https://www.elastic.co/docs/api/doc/elasticsearch/group/endpoint-connector), which defeats the purpose of declarative infrastructure and forces secrets onto operator workstations. Issue [#1572](https://github.com/elastic/terraform-provider-elasticstack/issues/1572) tracks the request.

## What Changes

- Add `elasticstack_elasticsearch_connector` resource with full CRUD over the connector envelope (`PUT /_connector/{id}`, `GET /_connector/{id}`, `DELETE /_connector/{id}`) plus the fan-out partial-update endpoints (`_name`, `_index_name`, `_service_type`, `_pipeline`, `_scheduling`, `_features`, `_api_key_id`, `_native`, `_configuration`).
- Model service-type-specific `configuration_values` as a typed nested map where each element selects exactly one branch (`string` / `number` / `bool` / `json` / `secret_value`), with `secret_value` marked write-only and protected against drift by the bcrypt-hash-in-private-state pattern.
- Add `data.elasticstack_elasticsearch_connector` companion data source surfacing the full read-time shape including runtime telemetry (`status`, `last_seen`, `last_synced`, `filtering`, `custom_scheduling`) that the resource deliberately omits.
- Add `elasticstack_elasticsearch_connector_sync_job_create` provider-defined action (Terraform 1.14+) that wraps `POST /_connector/_sync_job` for on-demand syncs, mirroring the existing `elasticstack_elasticsearch_snapshot_create` action pattern.
- Adopt the shared `internal/utils/writeonlyhash` helper for write-only secret drift detection. This change and the in-flight [fleet-cloud-connector change (PR #3415)](https://github.com/elastic/terraform-provider-elasticstack/pull/3415) both depend on the helper; whichever ships first builds it, the other consumes it.
- Add provider-wide documentation for `internal/utils/writeonlyhash` (package Godoc, usage examples, threat model, reference from `dev-docs/high-level/coding-standards.md`) regardless of which change ships first.

## Capabilities

### New Capabilities

- `elasticsearch-content-connector`: `elasticstack_elasticsearch_connector` resource and `data.elasticstack_elasticsearch_connector` data source covering the full connector envelope, fan-out updates, typed `configuration_values` map with per-element write-only secrets, and runtime-telemetry split between resource and data source.
- `elasticsearch-content-connector-sync-job`: `elasticstack_elasticsearch_connector_sync_job_create` provider-defined action that triggers on-demand sync jobs against an existing connector.
- `writeonly-secret-hashing-docs`: provider-wide documentation deliverable for the `internal/utils/writeonlyhash` helper — package Godoc, threat model, examples, and a reference in `dev-docs/high-level/coding-standards.md`. Built by this change regardless of which change builds the helper itself, so the docs land independently of helper-implementation ownership.

### Modified Capabilities

<!-- None. The writeonly-secret-hashing capability itself is defined by the fleet-cloud-connector change; this change consumes it and adds documentation under a separate capability rather than modifying the helper's requirements. -->

## Impact

- **New code**:
  - `internal/elasticsearch/connector/` — resource, data source, schema, models, fan-out client wrappers, validators, acceptance tests, examples.
  - `internal/elasticsearch/connector/sync_job_create/` — action implementation, schema, acceptance tests, examples.
  - `internal/clients/elasticsearch/connector.go` — thin wrappers over the typed `connector/*` packages from go-elasticsearch.
  - `dev-docs/high-level/writeonly-secret-hashing.md` — documentation for the shared helper; `coding-standards.md` updated to link to it.
- **Provider registration**: new resource, data source, and action registered in `provider/`.
- **Dependencies**:
  - Consumes `internal/utils/writeonlyhash` (built by this change or by [PR #3415](https://github.com/elastic/terraform-provider-elasticstack/pull/3415), whichever lands first).
  - Uses existing `github.com/elastic/go-elasticsearch/v8` typed `connector/*` packages — no new module dependencies.
- **Terraform version floor**: the action requires Terraform 1.14+ for provider-defined actions; the resource and data source have no new Terraform floor beyond what the provider already requires.
- **Elasticsearch version floor**: 8.12 (first GA of the connector APIs). Bump only if acceptance tests reveal blocking issues.
- **No breaking changes** to existing resources, data sources, or actions.
