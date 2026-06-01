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
| `entity_ids` update strategy | Set-diff: link new IDs, unlink removed IDs. The API supports partial ownership (unlinking a subset of IDs leaves the remaining group members intact), so `RequiresReplace` is unnecessary for `entity_ids`. |
| `resolution_group_json` | `jsontypes.NormalizedType{}` attribute (computed, normalized JSON). Populated by calling `GetSecurityEntityStoreResolutionGroup` with `entity_id = target_id` on every read. |
| Space routing | Use `kibanautil.SpaceAwarePathRequestEditor(spaceID)` as a request editor on all three API calls. The `default` space maps to an empty prefix (as per `BuildSpaceAwarePath`). |
| Read consistency | After link/unlink, the provider retries `GetSecurityEntityStoreResolutionGroup` with exponential back-off until the expected changes are visible in the response or a bounded timeout of approximately 2 seconds is reached, to account for the index refresh window. Acceptance tests may additionally use a small poll helper (mirroring patterns already used in Fleet resources) to wait for the API-observable state. |
| Import | `ImportState` parses `<space_id>/<target_id>` and reconstructs state from `GetSecurityEntityStoreResolutionGroup`. Since the raw resolution group is stored as `resolution_group_json`, all managed `entity_ids` can be reconstructed from the API response. |
| Data source `id` | Computed as `<space_id>/<entity_id>`. |
| Self-link validation | Schema validator at plan time: `target_id` MUST NOT appear in `entity_ids`. Enforce via `listvalidator.NoNullValues` + a custom `validator.List` that checks membership. |
| Version gate | `EnforceMinVersion("9.1.0")` in both Create and the data source Read callback. Versioned Elastic API docs for the resolution endpoints are present for current/v9 docs, while equivalent v8 versioned-doc URLs were not found during implementation, so the 9.1.0 gate remains unchanged unless acceptance testing proves otherwise. |
| Framework style | Plugin Framework (PF) — all new resources use PF, following `internal/kibana/agentbuilderagent/` and `internal/kibana/security_detection_rule/` as reference implementations. |

## Open Questions

1. **Set-diff update atomicity** — *Resolved*. The chosen strategy is set-diff (not `RequiresReplace`) because the API supports partial link/unlink operations. Unlinking a subset of `entity_ids` leaves the remaining resolution group members intact (confirmed by `POST /api/security/entity_store/resolution/unlink` semantics). Partial failure during an update is surfaced as an error diagnostic, and the next plan/apply will reconcile the remaining diff.

2. **Resolution group ownership**: The `delete` operation must only unlink the IDs managed by this resource, not the entire resolution group. Confirm that `POST /api/security/entity_store/resolution/unlink` with a subset of `entity_ids` leaves the remaining group members intact.

3. **Enterprise license skip strategy**: If the test environment does not have an enterprise license, link/unlink/group calls return a 403. The acceptance test suite must detect this and call `t.Skip()` rather than failing. Confirm the exact error shape from the API so a robust skip condition can be coded.

4. **Read consistency polling** — *Resolved*. The provider uses exponential back-off and retries the resolution-group GET until the expected changes are visible or a bounded timeout of approximately 2 seconds is reached. Acceptance tests may additionally use a small poll helper rather than a fixed sleep to wait for API-observable state and reduce flakiness.

5. **Minimum Stack version** — *Resolved*. Keep the minimum supported version at `9.1.0`. During implementation, the resolution endpoints were confirmed in current/v9 Elastic API docs, while equivalent v8 versioned-doc URLs returned not found, so there is no evidence of an 8.x backport from the documentation currently available. If acceptance testing later demonstrates supported 8.x behavior, the version gate can be relaxed in a follow-up change.

## Migration / State

- No existing resource covers this behavior; no state upgrade is required.
- Import path: `<space_id>/<target_id>` (e.g. `default/user-123`).
