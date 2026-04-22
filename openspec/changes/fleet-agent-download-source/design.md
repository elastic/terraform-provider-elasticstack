# Design: Fleet Agent Download Source requirements

## Context

Kibana Fleet exposes **Agent binary download sources** (`/api/fleet/agent_download_sources`) for configuring where Elastic Agent binaries are fetched. The provider already implements Terraform support under `internal/fleet/agentdownloadsource` and routes HTTP calls through `generated/kbapi` via `internal/clients/fleet`, consistent with other Fleet resources (e.g. outputs, server hosts).

This design records the technical decisions implied by the requirements and the existing implementation pattern.

## Goals / Non-Goals

**Goals:**

- Document API surface, client layering, and Terraform identity (`id` / `source_id`) for reviewers.
- Clarify space handling: `space_ids` selects the space for API calls; the API does not return space, so state preserves `space_ids`.
- Keep v1 scope limited: no Terraform attributes for Fleet `auth` and `secrets` on download sources.

**Non-Goals:**

- Defining new API behavior beyond what Kibana Fleet provides.
- Implementing `auth`/`secrets` in Terraform in this change.

## Decisions

### HTTP API surface

The resource SHALL use the Fleet endpoints for create, list, read single, update, and delete as documented in [Kibana Fleet agent download sources API](https://www.elastic.co/docs/api/doc/kibana/operation/operation-get-fleet-agent-download-sources). This matches REQ-style coverage in the delta spec.

### Client stack

All HTTP interactions SHALL go through the generated `kbapi` client and the Fleet wrapper in `internal/clients/fleet`, including space-aware request helpers when `space_ids` is set, matching other Fleet resources.

### Identity and import

- `id` in state SHALL store the Fleet download source ID returned by Kibana.
- `source_id` SHALL mirror that value for the path parameter on read, update, and delete.
- If Terraform sets `source_id` at create time, the create payload SHALL pass it so Kibana creates that ID; if omitted, the provider SHALL persist the server-assigned ID to `id` and `source_id`.
- Import SHALL accept a composite `<space_id>/<source_id>` identifier, set both `id` and `source_id` from `source_id`, and preserve `space_ids` as a single-entry set containing `space_id`.

### `space_ids`

Modeled as optional+computed `set(string)` (see `fleet-output` precedent). The first element determines the Kibana space for API calls; empty/unset means default space. `space_ids` SHALL NOT be overwritten from API responses.

### Errors

Non-success responses on mutating operations SHALL surface status and body in diagnostics. For read, `404` SHALL trigger removal from state; other errors SHALL be diagnostics.

### Post-mutation read convergence

After successful create and update operations, the provider SHALL perform a follow-up read using the same read helper/path as standard refresh. State SHALL be populated from the read response, not directly from create/update response bodies. This avoids state inconsistencies when mutation response shapes differ from read responses.

### Updates and replacement

In-place updates SHALL use `PUT` for `name`, `host`, `default`, and `proxy_id`. Changing `source_id` SHALL force replacement.

### Version guard

The provider SHALL enforce a minimum Kibana/Fleet version that supports this API (exact version TBD from product docs) and emit a clear diagnostic when the stack is too old.

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| API adds fields (e.g. auth) that Terraform does not model | Document in resource description; extend schema in a follow-up change |
| `space_ids` drift if user changes space outside Terraform | Same as other space-scoped Fleet resources; document that space is driven by config |

## Migration Plan

- Merge this change; implementers verify code against the delta spec.
- After implementation review, **sync** or **archive** so `openspec/specs/fleet-agent-download-source/spec.md` becomes canonical.

## Open Questions

1. **Minimum Kibana version** — Confirm exact version from Elastic release notes and encode in the version guard.
