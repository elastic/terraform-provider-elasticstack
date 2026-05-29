## Context

`elasticstack_kibana_space` uses `POST /api/spaces/space` for create and `DELETE /api/spaces/space/{id}` for destroy. Both calls fail unconditionally when targeting the default Kibana space (`space_id = "default"`):

- **Create**: Kibana returns 409 Conflict because the default space already exists on every deployment.
- **Destroy**: Kibana returns 400 Bad Request because the default space is a platform-level protected resource that cannot be deleted.

The current error surfacing is opaque: the practitioner sees raw HTTP status codes with no guidance.

Current implementation touch points:
- `internal/kibana/spaces/delete.go:28` — 3-line function, unconditional `kibanaoapi.DeleteSpace` call.
- `internal/clients/kibanaoapi/spaces.go:61-69` — `CreateSpace` function; 409 falls through to `HandleMutateTypedResponse`, which calls `diagutil.ReportUnknownHTTPError`.

## Goals / Non-Goals

**Goals:**

- Skip `DELETE /api/spaces/space/default` and remove from Terraform state cleanly, emitting `tflog.Warn` to surface the skip to operators.
- Return an actionable diagnostic on create 409 that points the practitioner to `terraform import`.
- Add an ungated acceptance test that imports the default space, updates it, and destroys without error.

**Non-Goals:**

- Auto-fallback from `POST 409` to `PUT` (auto-import). `terraform import` is the accepted workaround; the improved diagnostic covers the UX gap.
- Adding a `skip_delete` boolean schema attribute. The constraint is a platform invariant, not a user preference. Offering `skip_delete = false` as a valid default would be a footgun.
- Resetting the default space to factory configuration on destroy.
- Gating the acceptance test behind a stack version check.

## Decisions

| Topic | Decision | Alternatives considered |
|-------|----------|-------------------------|
| Destroy guard placement | Guard at the top of `deleteSpace` in `internal/kibana/spaces/delete.go`, before calling the API. | Placing the guard inside `kibanaoapi.DeleteSpace` was considered but rejected: `kibanaoapi` functions are generic API wrappers; the space-specific platform invariant belongs in the resource layer. |
| Guard condition | Hard-coded `resourceID == "default"`. | A `skip_delete` schema attribute (Approach B) was rejected by the maintainer. The constraint is a Kibana platform invariant, not a user choice. |
| Warning log | `tflog.Warn(ctx, "default Kibana space cannot be deleted; removing from Terraform state only")`. | Silently returning nil was rejected; surfacing intent to operators during `terraform destroy` is low cost and good practice. |
| Create 409 handler placement | Inside `kibanaoapi.CreateSpace`, before `HandleMutateTypedResponse`. | Placing it in the resource-layer `createSpace` function was considered; rejected because the HTTP status check is already done at the `kibanaoapi` layer, consistent with how other special statuses (404, 204) are handled in the same file. |
| Create 409 diagnostic | Error diagnostic naming the space id and instructing `terraform import elasticstack_kibana_space.<NAME> <id>`. | Returning a plain "space already exists" without import guidance was rejected as insufficiently actionable. |
| Acceptance test gating | Ungated (no version skip, no `solution` attribute). | Gating behind `>= 8.16.0` for `solution` was discussed and rejected by the maintainer: the fixture uses only `space_id` and `name`, which are supported on all stack versions. |
| Acceptance test destroy check | No `CheckDestroy`; verify that the destroy step itself completes without error. The default space persists after destroy — checking for its absence would always fail. | Using `CheckDestroy` was rejected because the space is never removed. |

## Risks / Trade-offs

- [Risk] The destroy skip is hardcoded to `space_id == "default"`. The constraint is a Kibana platform invariant; it is appropriate to encode it directly in the provider rather than make it configurable. No future version of Kibana allows deleting the default space.
- [Risk] The 409 diagnostic names the space id from the create request body. In theory, any POST conflict could hit this path. In practice, Kibana only 409s on `POST /api/spaces/space` when the id already exists, so the diagnostic is always accurate.
- [Trade-off] The acceptance test cannot assert that the space is gone after destroy (because it never is). The test instead asserts that the destroy step completes without error, which is the observable goal.

## Migration Plan

- No user-facing migration is required. Practitioners who were working around the 400 with `terraform import` can continue using the same flow; the destroy will now succeed cleanly.
- Practitioners who were hitting the 409 will see a helpful error message pointing to `terraform import` instead of an opaque "409 Conflict".
- No schema version or state upgrade is needed.

## Open Questions

- Non-blocking: Should the resource documentation explicitly note that the default space cannot be created via `terraform apply` and must be imported? This is a low-effort addition to `docs/resources/kibana_space.md` that would help operators before they hit the 409. Not blocked on this for implementation.
