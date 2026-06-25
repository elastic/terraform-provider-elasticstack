## Why

Kibana's Osquery integration lets teams group SQL queries into packs, schedule them via interval or RRULE-based schedules, and deploy them to Elastic Agent policies with optional per-policy shard percentages. Packs are the primary operational unit for fleet-wide scheduled osquery execution. Currently, the Terraform provider exposes no resource or data source for this surface, so teams must create and manage packs manually in Kibana UI or via out-of-band API calls, breaking IaC workflows and making multi-environment provisioning error-prone.

This change adds `elasticstack_kibana_osquery_pack` resource and matching data source, closing the gap for teams that want to declaratively manage Osquery packs alongside the rest of their Elastic Stack configuration.

## What Changes

- Add `elasticstack_kibana_osquery_pack` resource with full CRUD lifecycle against `/api/osquery/packs` (POST/GET/PUT/DELETE), composite import via `<space_id>/<pack_id>` (where `pack_id` is the server-generated `saved_object_id` UUID), and `kibana_connection` override consistent with other Kibana resources.
- Refuse to manage prebuilt (read-only) packs with a runtime error diagnostic guiding users to use the data source instead.
- Add `elasticstack_kibana_osquery_pack` data source (read-only GET-by-id) as the Terraform-native way to reference packs managed outside Terraform (including prebuilt packs).
- Model queries as a `MapNestedAttribute` keyed by query name with inline SQL, platform, ECS mapping, and other per-query options supported by the **pinned** kbapi client (v1 scope).
- Model `shards` as a `map(string → number)`; write as `map[string]float32`, read from GET map form, normalize create-response array quirk.
- **v1 scope (this change):** packs CRUD without scheduling fields — pinned kbapi lacks `schedule_type`, `rrule_schedule`, and per-query `interval`/`timeout` in generated types; kbapi bump blocked by `transform_schema.go`.
- **Deferred (follow-up):** dual scheduling modes and kbapi regeneration after Fleet transform fix.

## Capabilities

### New Capabilities

- `kibana-osquery-pack`: Defines the schema and runtime behavior of the `elasticstack_kibana_osquery_pack` resource, including identity (`pack_id` Computed from `saved_object_id`), space-awareness via `SpaceAwarePathRequestEditor`, prebuilt-pack guard, inline queries map, ECS mapping with three-way `field`/`value`/`values` exactly-one-of constraint, and `shards` normalization (v1 — no scheduling).
- `kibana-osquery-pack-datasource`: Defines the `elasticstack_kibana_osquery_pack` data source backed by GET-by-id, prebuilt-safe, enabling lookup of packs managed outside Terraform.

### Modified Capabilities

<!-- None — this change adds new capabilities only. -->

## Impact

- **New code**: `internal/kibana/osquery_pack/` (resource + data source), `internal/clients/kibanaoapi/osquery_pack.go` (thin client wrappers using generated kbapi bindings).
- **kbapi regeneration deferred**: bump blocked until `transform_schema.go` supports new Fleet response shapes; scheduling fields require post-bump follow-up.
- **New docs/examples**: `docs/resources/kibana_osquery_pack.md`, `docs/data-sources/kibana_osquery_pack.md`, `examples/resources/elasticstack_kibana_osquery_pack/`, `examples/data-sources/elasticstack_kibana_osquery_pack/`.
- **Provider registration**: register the new resource and data source in `provider/plugin_framework.go`.
- **Minimum version**: `8.5.0` for v1 packs CRUD.
- **Backward compatibility**: additive only — no breaking changes to existing resources or data sources.
