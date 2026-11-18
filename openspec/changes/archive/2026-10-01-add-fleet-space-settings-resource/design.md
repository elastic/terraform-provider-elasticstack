## Context

See proposal.md for motivation and specs/fleet-space-settings/spec.md for requirements.

Current state and constraints:

- The generated Kibana client already exposes the Fleet space settings operations (`GetFleetSpaceSettings`, `PutFleetSpaceSettings`); the request body has a single optional `allowed_namespace_prefixes` array, the response wraps `{ item: { allowed_namespace_prefixes, managed_by? } }`.
- The Kibana API documentation states that the operation was added in 9.1.0, that a write accepts at most 10 prefixes, and that a response contains at most 100 (https://www.elastic.co/docs/api/doc/kibana/operation/operation-put-fleet-space-settings); the 10-element limit and the 9.1.0 minimum in the spec derive from this.
- The settings object always exists for a space; there is no create or delete endpoint, only read and write.
- Existing Fleet resources (`serverhost`, `proxy`, `output`, ...) are built on `entitycore.KibanaResource[T]`, which provides `kibana_connection`, timeouts, version-requirement enforcement, read-after-write and not-found handling. Version requirements are enforced by the envelope on every lifecycle operation, including read and delete.
- Fleet API calls are routed to a space with the space-aware path request editor used by the existing Fleet client wrappers.

## Goals / Non-Goals

**Goals:**
- Implement the resource with the same envelope and package layout as other Fleet resources so that connection handling, timeouts and version gating behave identically.
- Keep all state derived from the Fleet API response (no provider-side state beyond the identity).

**Non-Goals:**
- No changes to `entitycore` (for example relaxing the version check on delete).
- No shared abstraction for "singleton per space" resources; this is the first such Fleet resource and a generalisation would be speculative.
- No cross-resource coordination for multiple Terraform resources targeting the same space.

## Decisions

### Use the `entitycore.KibanaResource` envelope

The resource is a new package `internal/fleet/spacesettings` embedding `entitycore.KibanaResource[model]`, registered in the provider as `elasticstack_fleet_space_settings`.

Rationale: matches `serverhost`/`proxy` and the repo coding standards; connection override, timeouts, version gate and not-found handling come for free.
Alternative considered: a standalone resource with hand-written CRUD. Rejected because it duplicates envelope behaviour and diverges from repo conventions.

### Components

| Unit | Responsibility |
|---|---|
| `schema.go` | `space_id` is a required string with a replace-on-change modifier, defined locally because the shared `kbschema` space attributes are optional-with-default; `allowed_namespace_prefixes` is a required set of strings with a maximum size of 10 (no explicit uniqueness validator, since a set's elements are inherently unique); `managed_by` and `id` are computed. |
| `models.go` | Model exposing `space_id` as the resource ID, the write ID and the space for envelope routing. Provides the version requirement (minimum 9.1.0) and mapping between the API item and the model, mapping a missing API list to an empty set. |
| `create.go` | A single write callback, used for both create and update (no separate `update.go`), that sends the planned set in one PUT for the planned space. |
| `read.go` | GET for the space; returns "not found" when the client wrapper returns nil. |
| `delete.go` | PUT with an empty set for the space; a not-found response is treated as success. |
| `internal/clients/fleet/space_settings.go` | Wrapper functions `GetSpaceSettings` / `PutSpaceSettings` / `ResetSpaceSettings` (PUT of an empty set where a 404 counts as success, used by destroy) using the existing typed-response helpers and space-aware path editor; GET returns nil on HTTP 404. |

Import: the import ID is the space ID and is written to both `id` and `space_id`, after which the standard read populates the rest.

### Identity is the space ID

`id` and the envelope's resource ID both equal `space_id`. There is exactly one settings object per space, so no composite ID is needed. `space_id` forces replacement.

### Data flow

Plan → write callback issues a single PUT (space in the request path) → the envelope performs a read-after-write GET → state is built exclusively from the GET response: `allowed_namespace_prefixes` is a set, so order carries no meaning; a missing list in the response becomes an empty set, `managed_by` is null when absent. Destroy issues a PUT with an empty set and drops the resource from state.

### Failure handling

- Read: HTTP 404 removes the resource from state; other errors become diagnostics.
- Destroy: HTTP 404 counts as success so that a deleted space does not block destroy.
- Create/Update: API errors (for example a missing `fleet-settings-all` privilege) are surfaced unchanged; the resource does not check that the space exists, because Kibana 9.4 accepts settings for a non-existent space (documented as a risk); no retry logic is added because the PUT is idempotent and no conflict handling is documented for this endpoint.
- Version too old: the envelope's version requirement fails the operation before any settings API call.
- Plan-time validation (more than 10 elements) prevents API calls entirely.

### Testing approach

- Unit tests for the model/API mapping (empty, missing list, present and absent `managed_by`), the schema's size validator (11 elements) and the version requirement.
- Acceptance tests in the package against a stack of at least 9.1: create, update, import by space ID, destroy resets the set while the space remains, drift after an out-of-band change, and a non-default space (created with `elasticstack_kibana_space`). The default-space import test writes into the shared `default` space of the test stack and relies on destroy to reset it. Tests follow the repo's testing docs and skip on older stack versions. Because `allowed_namespace_prefixes` is a set, the legacy acceptance-testing state representation exposes elements as hash-keyed flatmap entries rather than positional indices, so element checks use `TestCheckTypeSetElemAttr` (`attr.*`) instead of `attr.0`/`attr.1`.

## Risks / Trade-offs

- [A mistyped or non-existent `space_id` silently creates orphan settings, because Kibana does not reject them] → Document the behavior; no provider-side existence check (see proposal non-goals for scope).
- [Two Terraform resources target the same space and overwrite each other's prefixes] → Document that exactly one `elasticstack_fleet_space_settings` should exist per space; no provider-side detection.
- [Destroy removes a security control] → Document that destroy resets the set to empty and therefore lifts the namespace restriction.
- [The envelope enforces the version requirement on read and delete, so an older server blocks destroy] → Accepted: the resource can never have been created on a server older than 9.1.0, so a state entry against such a server is not expected in practice.
- [Kibana may reorder or normalise prefixes] → Not a risk: `allowed_namespace_prefixes` is a set, so element order is never significant and cannot cause drift.
- [Values above the 10-element write limit already stored in Kibana (up to 100 readable) cannot be written back by Terraform] → Read shows them as drift; the plan-time validation blocks applying an oversized set. No workaround is provided (see proposal non-goals).

## Migration Plan

Not applicable: this is a new resource with no impact on existing configurations. Rollback is removing the resource from configuration, which resets the space's prefixes to empty on destroy.
