## Context

`elasticstack_kibana_synthetics_parameter` currently routes every CRUD call through `kibanaClient.API` without a `RequestEditorFn`, so all operations hit `/api/synthetics/params[/{id}]` in whatever space the Kibana client is authenticated to — typically the default space. The model implements `KibanaUnscopedSpace` (`IsUnscopedSpace() bool { return true }`) to suppress the envelope's empty-space-id check, and `GetSpaceID()` returns `""`.

The Kibana Synthetics Parameters API is fully space-aware; the space prefix (`/s/{space_id}`) is injected before `/api/synthetics/params` by `kibanautil.SpaceAwarePathRequestEditor`. This helper is already used by the monitor resource.

The existing `id` field is set to the bare UUID returned by Kibana on create. `resolveKibanaResourceIdentity` (in `entitycore`) parses `GetID()` as a composite `<cluster>/<resource>` first; if that fails it falls back to `GetResourceID()`/`GetSpaceID()`. The parameter model satisfies both branches today because `GetSpaceID()` returns `""` and `IsUnscopedSpace()` is true — the envelope skips the non-empty-space-id guard.

## Goals

1. Add `space_id` (`optional, computed`, `UseStateForUnknown`, `RequiresReplace`) to the parameter resource schema.
2. Store `id` as `<space_id>/<parameter_uuid>` so `resolveKibanaResourceIdentity` recovers space from state without needing a separate `space_id` field on read.
3. Pass `kibanautil.SpaceAwarePathRequestEditor(spaceID)` in all four CRUD operations so the Kibana API path is space-prefixed when `space_id` is non-default.
4. Remove `KibanaUnscopedSpace` / `IsUnscopedSpace()` from the model; the envelope will validate that `space_id` is a known, non-empty value on create/update.
5. Add `StateUpgraders` v0→v1 to migrate existing state (bare UUID `id`) to `default/<uuid>` and add `space_id = "default"`.
6. Update `ImportState` to accept both bare UUID (maps to default space) and `<space_id>/<uuid>`.

## Non-Goals

- Adding space support to `elasticstack_kibana_synthetics_private_location`.
- Changing `share_across_spaces` semantics.
- Bulk/batch parameter import.

## Decisions

| Topic | Decision |
|-------|----------|
| `space_id` schema | `optional, computed`, `UseStateForUnknown`, `RequiresReplace` — mirrors `elasticstack_kibana_synthetics_monitor`. No default; `"default"` is inferred by `SpaceAwarePathRequestEditor` when the value is empty. |
| Composite `id` | Store as `<space_id>/<parameter_uuid>` (same as monitor). `modelFromOAPI` must accept the space from the write request, not re-derive it from the API response (the API does not echo space). |
| `KibanaUnscopedSpace` removal | Remove `IsUnscopedSpace()` and the `_ entitycore.KibanaUnscopedSpace = Model{}` assertion. The envelope will enforce a non-empty `space_id` going forward. |
| `GetSpaceID()` | Return `m.SpaceID` (the typed `types.String`). Return `""` when null/unknown to preserve the default-space routing behavior. |
| `GetResourceID()` | Parse `GetID()` as composite; return UUID segment. Fallback to bare `id` for v0 legacy state during migration window. |
| State migration | Schema version bump v0→v1. `UpgradeState` v0 handler rewrites `id` to `default/<id>` and adds `space_id = "default"`. |
| Import | Accept `<space_id>/<uuid>` or bare `<uuid>`; bare maps to default space (sets `space_id = "default"`). Implement via `ImportState` override on the `Resource` struct rather than `ImportStatePassthroughID`. |
| CRUD routing | Append `kibanautil.SpaceAwarePathRequestEditor(spaceID)` to each `*WithResponse` call (same helper used by monitor's `kibanaoapi.CreateMonitor`). |
| Read populates `SpaceID` | `modelFromOAPI` must be extended to accept a `spaceID string` argument; it stores both the composite `id` and the `space_id` field from the write context. On plain reads, space is recovered from `resolveKibanaResourceIdentity` via the composite `id`. |

## Risks / Trade-offs

- **State-breaking change**: existing users with default-space parameters will have their `id` rewritten from a bare UUID to `default/<uuid>` on the first `terraform apply` after upgrade (triggered by the state upgrader). This is safe — no destroy/recreate — but users should be informed in the changelog.
- **Non-default `space_id` + `share_across_spaces = true`**: the Kibana API does not prevent this combination; the provider follows the same non-validating approach as the monitor resource. This should be documented but not enforced at the provider level.

## Open Questions

- **Version gate**: The current resource gates on Kibana 8.12.0. Is the space-prefixed path available from the same version, or does it require a later release? _Assumption_: space-scoped Synthetics Parameters routes were available at least from 8.12.0 (same as the unscoped path). If a later gate is required, an additional `EnforceMinVersion` call should be added in each CRUD operation that uses the space prefix. An implementer should verify this against Kibana release notes or the API changelog.
- **`share_across_spaces` + non-default `space_id`**: Should the provider validate or document creating a parameter in `my-space` with `share_across_spaces = true`? _Assumption_: document only; do not validate. Kibana may allow it (shared from a non-default space). Follow up if QA reveals API rejection.
