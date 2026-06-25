## Why

Kibana's Osquery integration lets teams group SQL queries into packs, schedule them via interval or RRULE-based schedules, and deploy them to Elastic Agent policies with optional per-policy shard percentages. Packs are the primary operational unit for fleet-wide scheduled osquery execution. Currently, the Terraform provider exposes no resource or data source for this surface, so teams must create and manage packs manually in Kibana UI or via out-of-band API calls, breaking IaC workflows and making multi-environment provisioning error-prone.

This change adds `elasticstack_kibana_osquery_pack` resource and matching data source, closing the gap for teams that want to declaratively manage Osquery packs alongside the rest of their Elastic Stack configuration.

## What Changes

- Add `elasticstack_kibana_osquery_pack` resource with full CRUD lifecycle against `/api/osquery/packs` (POST/GET/PUT/DELETE), composite import via `<space_id>/<pack_id>`, and `kibana_connection` override consistent with other Kibana resources.
- Refuse to manage prebuilt (read-only) packs with a runtime error diagnostic guiding users to use the data source instead.
- Add `elasticstack_kibana_osquery_pack` data source (read-only GET-by-id) as the Terraform-native way to reference packs managed outside Terraform (including prebuilt packs).
- Model queries as a `MapNestedAttribute` keyed by query name with inline SQL, scheduling overrides, platform, ECS mapping, and other per-query options.
- Support dual scheduling modes at both the pack level and per-query override level: `interval` (seconds, Int64) and `rrule_schedule` (RFC 5545 RRULE with `start_date`, optional `end_date`, `splay`, `timeout`).
- Model `shards` as a `map(string → number)` normalized from the API's array-on-create / map-on-read format difference.
- Regenerate the `kbapi` client (bump OAS ref) to include the modern scheduling fields (`schedule_type`, `rrule_schedule`, per-query `interval`/`timeout`/`splay`) missing from the pinned client.

## Capabilities

### New Capabilities

- `kibana-osquery-pack`: Defines the schema and runtime behavior of the `elasticstack_kibana_osquery_pack` resource, including identity (`pack_id` Optional+Computed+RequiresReplace), space-awareness via `SpaceAwarePathRequestEditor`, prebuilt-pack guard, inline queries map with per-query scheduling overrides, ECS mapping with three-way `field`/`value`/`values` exactly-one-of constraint, dual scheduling mode validation, and `shards` normalization.
- `kibana-osquery-pack-datasource`: Defines the `elasticstack_kibana_osquery_pack` data source backed by GET-by-id, prebuilt-safe, enabling lookup of packs managed outside Terraform.

### Modified Capabilities

<!-- None — this change adds new capabilities only. -->

## Impact

- **New code**: `internal/kibana/osquery_pack/` (resource + data source), `internal/clients/kibanaoapi/osquery_pack.go` (thin client wrappers using generated kbapi bindings).
- **kbapi regeneration required**: bump OAS ref in `generated/kbapi/Makefile` to a version that includes `schedule_type`/`rrule_schedule`/per-query `interval`/`timeout`/`splay`; run `make -C generated/kbapi all`. Verify during implementation that the bumped ref includes these fields.
- **New docs/examples**: `docs/resources/kibana_osquery_pack.md`, `docs/data-sources/kibana_osquery_pack.md`, `examples/resources/elasticstack_kibana_osquery_pack/`, `examples/data-sources/elasticstack_kibana_osquery_pack/`.
- **Provider registration**: register the new resource and data source in `provider/plugin_framework.go`.
- **Minimum version**: `8.5.0` for base packs CRUD; a higher floor (TBD during implementation) for `schedule_type`/`rrule_schedule` scheduling fields via `GetVersionRequirements`.
- **Backward compatibility**: additive only — no breaking changes to existing resources or data sources.
