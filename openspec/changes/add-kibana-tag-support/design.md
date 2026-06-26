## Context

Kibana [PR elastic/kibana#266220](https://github.com/elastic/kibana/pull/266220) introduced a public REST API for Kibana tags at `/api/tags`, shipping in Kibana 9.5.0. The legacy `/api/saved_objects_tagging/tags/...` endpoints are now deprecated. The generated `kbapi` client already models the full surface:

| Operation | Endpoint               | kbapi method         | Request type                              |
|-----------|------------------------|----------------------|-------------------------------------------|
| List      | `GET /api/tags`        | `GetTags`            | `GetTagsParams` (query, page, per_page)   |
| Get       | `GET /api/tags/{id}`   | `GetTagsId`          | path param `id`                           |
| Create    | `POST /api/tags`       | `PostTags`           | `KibanaHTTPAPIsKbnTagsRequestAttributes`  |
| Upsert    | `PUT /api/tags/{id}`   | `PutTagsId`          | `KibanaHTTPAPIsKbnTagsRequestAttributes`  |
| Delete    | `DELETE /api/tags/{id}`| `DeleteTagsId`       | path param `id`                           |

Response shape (all single-tag endpoints):
- `id` (string) — server-generated UUID
- `data` (`KibanaHTTPAPIsKbnTagsAttributes`): `name` (string), `color` (string), `description` (*string)
- `meta` (`KibanaHTTPAPIsKbnAsCodeMeta`): `created_at`, `created_by`, `updated_at`, `updated_by`, `managed` (*bool), `owner`, `version`

`PutTagsId` returns `JSON200` on update and `JSON201` on upsert-create (tag did not exist before the PUT).

## Goals

- Provide `elasticstack_kibana_tag` resource for full CRUD lifecycle management of a single Kibana tag.
- Support both server-minted IDs (POST) and client-specified IDs (PUT-to-create) for importing existing infrastructure.
- Guard against Kibana-managed tags (those with `meta.managed = true`) — the resource refuses to read, update, or delete them.
- Provide `elasticstack_kibana_tags` data source for listing tags by name/description query, suitable for reference in downstream resources.
- Gate both entities on Kibana ≥ 9.5.0 with a clear diagnostic.

## Non-Goals

- Supporting the deprecated `/api/saved_objects_tagging/tags/...` endpoints.
- Managing tag assignments to saved objects (dashboards, visualizations, etc.) — that is a concern of those resources.
- Exposing `meta.version` for optimistic concurrency control (no other PF resource in this repo wires it into `If-Match`).
- Exposing `meta.managed`, `meta.owner`, `meta.created_by`, `meta.updated_by` as schema attributes — these are server-internal.

## Decisions

| Topic | Decision |
|-------|----------|
| Framework | Plugin Framework (terraform-plugin-framework), consistent with all recent Kibana resources in this repo |
| Resource name | `elasticstack_kibana_tag` (singular, consistent with `elasticstack_kibana_osquery_pack`) |
| Data source name | `elasticstack_kibana_tags` (plural list, consistent with `elasticstack_kibana_spaces`) |
| `tag_id` (optional, ForceNew) | Allows practitioners to specify a deterministic UUID; maps directly to the `{id}` path parameter on `PUT`. When unset, provider uses POST and stores the server-minted UUID. ForceNew because the tag UUID is the stable identity key. |
| `color` Optional+Computed | When omitted in config, server generates a random color; `UseStateForUnknown` prevents spurious diffs after first create. Removing `color` from HCL after the initial apply will not cause a diff (consistent with computed-when-unset semantics). |
| `space_id` Optional+Computed ForceNew | Defaults to `"default"`. Space-awareness is handled via the KibanaAPI client's space header/path mechanism already established in the provider. ForceNew because a tag's space is immutable after creation. |
| Composite `id` | `"<space_id>/<tag_id>"` — consistent with osquery pack, SLO, and other space-aware resources. |
| Create branching | `tag_id` absent → `POST /api/tags`; `tag_id` present → `GET /api/tags/{id}`: 404 → `PUT` to create; 200 → error diagnostic directing user to `terraform import`. This prevents silent adoption of an existing tag. |
| Update | Straight `PUT /api/tags/{id}`. The API's upsert semantics (`JSON201`) are treated as a no-op race condition guard — the steady-state external-delete case is handled by Terraform's pre-plan refresh. |
| Managed tag guard | On Read, Update, and Delete, if `meta.managed = true`, return an error diagnostic and do not modify state. Pattern mirrors `internal/kibana/osquery_pack/guard.go`. The resource cannot legitimately produce a managed tag, so `managed` is not exposed in the schema. |
| Import | `"<space_id>/<id>"` format parsed to extract space and tag ID. Import refuses managed tags (guard runs on the post-import read). |
| `created_at` / `updated_at` | Stored as Computed strings from `meta.created_at` / `meta.updated_at`. Not user-settable. |
| Data source pagination | Auto-paginates: fetches pages until `len(collected) >= meta.total`. Uses `per_page=100` (or server max if lower) to minimize round trips. |
| Data source `query` | Passed verbatim as `GetTagsParams.Query`; supports full Elasticsearch `simple_query_string` syntax (e.g. `+name:prod -description:legacy`). Empty/absent → no filter. |
| Version gate | Check at resource/data source entry points; use existing version-check helper. Minimum: `9.5.0`. Return an `ErrorDiagnostic` with a message pointing to the Kibana 9.5 release if the check fails. |

## Non-Goals (implementation)

- Do not implement tag assignment to other Kibana saved objects in this change.
- Do not expose `meta.version` for optimistic locking.

## Risks / Trade-offs

- **Color drift**: Kibana stores colors as provided. If a practitioner omits `color` after initial create, `UseStateForUnknown` prevents a plan diff. If they later set `color` to a different value in config, Terraform will show a diff and apply the update correctly. No known issue.
- **Name uniqueness**: Kibana's behavior on duplicate tag names is not confirmed from release notes. The API likely enforces uniqueness per space. If a 4xx is returned on create with a conflicting name, the provider will surface the raw Kibana error diagnostic. Name uniqueness rules, if confirmed during implementation, should be added as validator.
- **Managed tag after import**: Import runs a Read, which triggers the managed guard. Practitioners who try to import a managed tag will receive an error with a useful message.

## Open Questions

1. **Name validation**: Does Kibana enforce a maximum name length or character constraints? If confirmed during implementation (e.g., via API 4xx responses), add a `stringvalidator.LengthBetween` or similar validator.
2. **Color validation**: Is the hex color format strictly enforced (must be `#RRGGBB`), or does Kibana accept short-form (`#RGB`)? Confirm during acceptance testing; add `stringvalidator.RegexMatches` if strict.
3. **Pagination max per_page**: Does the server cap `per_page`? If so, the data source auto-pagination must not exceed the cap. Investigate during implementation; default to 100 and handle server rejection gracefully.
4. **Name uniqueness enforcement**: Server-side uniqueness per space — error response shape and HTTP status to be confirmed during implementation.
5. **`description` null vs empty-string round-trip**: The API models `description` as `*string` (optional). It is not yet confirmed whether sending `description = ""` round-trips identically to sending `description: null` (omitted), or whether the server normalizes empty to absent. If the two are not equivalent, Terraform could observe a perpetual diff when a practitioner writes `description = ""` in HCL. Confirm during implementation via a round-trip unit test on `toAPIModel`/`fromAPIModel`; add a normalizer (treat empty string as absent on write) if necessary.

## Migration / State

- These are new resources with no prior state. No state upgrader is required.
- Practitioners currently using `elasticstack_kibana_import_saved_objects` for tags can migrate by importing the existing tag UUIDs via `terraform import elasticstack_kibana_tag.<name> <space_id>/<tag_id>`.
