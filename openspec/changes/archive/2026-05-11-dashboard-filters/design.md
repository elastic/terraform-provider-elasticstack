## Context

The Kibana Dashboard API persists saved filter pills under `kbn-dashboard-data.filters` as a list of objects discriminated by `operator` (or DSL/spatial shapes for the more exotic branches). The TF resource currently exposes per-panel `filter_json` lists of the same shape but no dashboard-level equivalent.

## Goals / Non-Goals

**Goals:**
- Round-trip dashboard-level saved filters with semantic JSON equality.
- Stay consistent with the existing per-panel `filter_json` pattern so practitioners learn one shape.
- Avoid locking the schema to the current discriminator set; the API can grow new operator variants without forcing a TF schema change.

**Non-Goals:**
- Typed/structured filter blocks (one nested attribute per `operator`). Punted: the discriminator surface is wide and Kibana-generated; users typically copy-paste saved-filter JSON from the Kibana UI's "view as JSON" tooling.
- Filter validation beyond well-formed JSON. The API performs its own validation; surfacing a richer validator here doubles maintenance.

## Decisions

- **Shape**: `filters = list(object({ filter_json = string }))`. Mirrors the existing per-panel block exactly. Each list element is one filter; the wrapping object is retained for forward extensibility (e.g., adding optional `disabled` or `pinned` flags later without breaking the schema).
  - *Rejected*: bare `filters_json = string` (single JSON list). Loses item-level diffing; mass-replaces the whole list whenever any filter changes.
  - *Rejected*: structured discriminated nested blocks (`is = {...}`, `is_one_of = {...}`, …). High maintenance cost, brittle against Kibana spec evolution, no immediate user-visible benefit over JSON.
- **Normalization**: reuse the existing `config_json` semantic-equality plan modifier so key reorderings and Kibana-injected defaults do not appear as diffs.
- **Null vs empty**: an unset `filters` attribute stays unset on read when the API returns either no `filters` field or an empty list (consistent with REQ-009 "State preservation for fields Kibana omits or defaults").
- **Generated client gap (`PutDashboardsIdJSONBody_Filters_Item`)**: the OpenAPI-generated PUT-body filter union type ships without `MarshalJSON`/`UnmarshalJSON`, unlike `KbnDashboardData_Filters_Item`, so `encoding/json` cannot round-trip the union from other packages. The hand-maintained companion `generated/kbapi/put_dashboard_filters_item_json.go` adds those methods; **delete that file if a future `make generate` ever emits equivalent methods** (otherwise the build will fail with duplicate receivers).

## Risks / Trade-offs

- [Risk] Practitioners may author malformed filter JSON that Kibana later rejects → Mitigation: validator ensures well-formed JSON at plan time; runtime errors surface from the API call as today.
- [Risk] Future API changes to the filter discriminator are silently accepted → Mitigation: this is intentional (forward-compatible); the JSON-first stance is the same trade-off the per-panel filters made.
