## Why

Kibana's Osquery integration allows teams to run scheduled SQL queries against endpoints managed by Elastic Agent. Saved queries are reusable query definitions that can be referenced by detection rules (via `response_actions[].params.saved_query_id`) and Osquery packs. Currently, the Terraform provider exposes no resource or data source for this surface, so teams must create and manage these queries manually in Kibana UI or rely on out-of-band API calls, breaking IaC workflows and making multi-environment provisioning (dev/staging/prod) error-prone.

This change adds `elasticstack_kibana_osquery_saved_query` resource and matching data source, closing the gap for teams that want to declaratively manage Osquery saved queries alongside the rest of their Elastic Stack configuration.

## What Changes

- Add `elasticstack_kibana_osquery_saved_query` resource with full CRUD lifecycle against `/api/osquery/saved_queries` (POST/GET/PUT/DELETE), composite state/import `id` (`<space_id>/<saved_query_id>`) with API lookup via `saved_query_id`, and `kibana_connection` override consistent with other Kibana resources.
- Refuse to manage prebuilt queries (those with `prebuilt == true` in the API response) with a runtime error diagnostic guiding users to use the data source instead.
- Add `elasticstack_kibana_osquery_saved_query` data source (read-only GET-by-id) as the only Terraform-native way to reference prebuilt queries or queries managed outside Terraform (e.g., ones referenced by `response_actions[].params.saved_query_id` in a detection rule).
- Model `ecs_mapping` as a `MapNestedAttribute` of `SingleNestedAttribute` with `field`, `value`, and `values` fields — matching the `{Field, Value: string|[]string}` union returned by the generated `kbapi` client.
- Normalise `interval` and `version` from kibanaoapi entity types to `int64` and `string` in state. Create and GET entities use `json.RawMessage` unions (`AsXxx0()/AsXxx1()` accessors); the Update entity types `version` as plain `*string` while `interval` remains a union.

## Capabilities

### New Capabilities

- `kibana-osquery-saved-query`: Defines the schema and runtime behavior of the `elasticstack_kibana_osquery_saved_query` resource, including composite state `id` (`<space_id>/<saved_query_id>`), API lookup via `saved_query_id` (Required+RequiresReplace), space-awareness via `SpaceAwarePathRequestEditor`, prebuilt-query guard, ECS mapping with three-way `field`/`value`/`values` exactly-one-of constraint via `ExactlyOneOfNestedAttrsValidator`, platform comma-string wire format, and `interval`/`version` response normalisation.
- `kibana-osquery-saved-query-datasource`: Defines the `elasticstack_kibana_osquery_saved_query` data source backed by GET-by-id, prebuilt-safe, with the same `8.5.0` version floor via `GetVersionRequirements`.

### Modified Capabilities

<!-- None — this change adds new capabilities only. -->

## Impact

- **New code**: `internal/kibana/osquery_saved_query/` (resource + data source), `internal/clients/kibanaoapi/osquery_saved_query.go` (thin client wrappers using existing generated kbapi bindings).
- **No kbapi regeneration needed**: all four CRUD bindings already exist in `generated/kbapi/kibana.gen.go` (`OsqueryCreateSavedQuery`, `OsqueryGetSavedQueryDetails`, `OsqueryUpdateSavedQuery`, `OsqueryDeleteSavedQuery`). Space support is injected via `kibanautil.SpaceAwarePathRequestEditor` (no `transform_schema.go` changes needed).
- **New docs/examples**: `docs/resources/kibana_osquery_saved_query.md`, `docs/data-sources/kibana_osquery_saved_query.md`, `examples/resources/elasticstack_kibana_osquery_saved_query/`, `examples/data-sources/elasticstack_kibana_osquery_saved_query/`.
- **Provider registration**: register the new resource and data source in `provider/plugin_framework.go`.
- **Minimum version**: `8.5.0` documented/conservative floor from Kibana API docs and source (task 1.2); `GetVersionRequirements` in task 3.2; live confirmation deferred to acceptance task 9.3.
- **Backward compatibility**: additive only — no breaking changes to existing resources or data sources.
