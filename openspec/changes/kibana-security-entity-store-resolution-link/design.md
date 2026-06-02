## Context

The Kibana Entity Store resolution API (introduced in Elastic Stack 9.1.0) allows linking multiple alias entity identifiers to a single "golden" target entity, forming a resolution group. Changes are eventually consistent (index refresh, typically < 1 s).

API operations available in `generated/kbapi/kibana.gen.go`:

| Operation | Method | Path | kbapi type |
|-----------|--------|------|------------|
| Link | POST | `/api/security/entity_store/resolution/link` | `PostSecurityEntityStoreResolutionLink` |
| Unlink | POST | `/api/security/entity_store/resolution/unlink` | `PostSecurityEntityStoreResolutionUnlink` |
| Get group | GET | `/api/security/entity_store/resolution/group?entity_id=<id>` | `GetSecurityEntityStoreResolutionGroup` |

Space routing is achieved via `kibanautil.SpaceAwarePathRequestEditor(spaceID)` passed as a `RequestEditorFn`.

Response bodies for all three operations are returned as raw `[]byte` in the current generated client (no typed `JSON200` fields). The resource and data source must unmarshal these manually.

## Goals

- Provide a Terraform resource (`elasticstack_kibana_security_entity_store_entity_link`) that manages the full lifecycle of an entity resolution link.
- Provide a Terraform data source (`elasticstack_kibana_security_entity_store_resolution_group`) that reads the resolution group for any entity.
- Enforce schema validation at plan time (entity_ids size, self-link guard).
- Support import via composite ID `<space_id>/<target_id>`.

## Non-Goals

- Creating or deleting entities in the Entity Store (out of scope; use a separate entity resource).
- Managing Entity Store engine lifecycle (install/uninstall/start/stop).
- Typed response structs in `generated/kbapi/` — implementation parses raw JSON responses.

## Decisions

| Topic | Decision |
|-------|----------|
| Resource ID | Composite `<space_id>/<target_id>`. Stable; does not encode `entity_ids` because the set can be updated in place. |
| `entity_ids` update strategy | Set-diff: link new IDs, unlink removed IDs. If API behavior makes this brittle (e.g. Kibana rejects partial ownership), add `RequiresReplace` for `entity_ids` as a fallback and document in the delta spec open questions. |
| `resolution_group_json` | `jsontypes.NormalizedType{}` attribute (computed, normalized JSON). Populated by calling `GetSecurityEntityStoreResolutionGroup` with `entity_id = target_id` on every read. |
| Space routing | Use `kibanautil.SpaceAwarePathRequestEditor(spaceID)` as a request editor on all three API calls. The `default` space maps to an empty prefix (as per `BuildSpaceAwarePath`). |
| Read consistency | After link/unlink, retry the GET with exponential back-off (up to ~2 s total) to account for the index refresh window. A small poll helper (mirroring patterns already used in Fleet resources) should be used in acceptance tests. |
| Import | `ImportState` parses `<space_id>/<target_id>` and reconstructs state from `GetSecurityEntityStoreResolutionGroup`. Since the raw resolution group is stored as `resolution_group_json`, all managed `entity_ids` can be reconstructed from the API response. |
| Data source `id` | Computed as `<space_id>/<entity_id>`. |
| Self-link validation | Schema validator at plan time: `target_id` MUST NOT appear in `entity_ids`. Enforce via `listvalidator.NoNullValues` + a custom `validator.List` that checks membership. |
| Version gate | `EnforceMinVersion("9.1.0")` in both Create and the data source Read callback. Adjust if acceptance testing reveals a different supported version. |
| Framework style | Plugin Framework (PF) — all new resources use PF, following `internal/kibana/agentbuilderagent/` and `internal/kibana/security_detection_rule/` as reference implementations. |

## Open Questions

1. **Set-diff update atomicity**: The API accepts batches of up to 1000 entity IDs per call. If a link fails partway through a diff (e.g. network error after linking new IDs but before unlinking old ones), the state may be partially applied. Consider wrapping the diff in a single `RequiresReplace` policy for `entity_ids` initially and graduating to set-diff only after acceptance-test validation confirms idempotent behavior.

2. **Resolution group ownership**: The `delete` operation must only unlink the IDs managed by this resource, not the entire resolution group. Confirm that `POST /api/security/entity_store/resolution/unlink` with a subset of `entity_ids` leaves the remaining group members intact.

3. **Enterprise license skip strategy**: If the test environment does not have an enterprise license, link/unlink/group calls return a 403. The acceptance test suite must detect this and call `t.Skip()` rather than failing. Confirm the exact error shape from the API so a robust skip condition can be coded.

4. **Read consistency polling**: The Entity Store API documentation states changes are visible after the next index refresh (< 1 s). Acceptance tests that create-then-read without a poll may be flaky. Decide on the polling strategy (busy-loop with timeout vs. `time.Sleep`) before implementing acceptance tests.

5. **Minimum Stack version**: The issue specifies 9.1.0 as the starting point. Verify that the resolution API was not backported to 8.x; if it was, adjust the version gate accordingly.

## Migration / State

- No existing resource covers this behavior; no state upgrade is required.
- Import path: `<space_id>/<target_id>` (e.g. `default/user-123`).
