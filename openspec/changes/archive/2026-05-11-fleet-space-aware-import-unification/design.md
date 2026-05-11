## Context

Six Fleet resources implement `resource.ResourceWithImportState`. Two (`fleet_output`, `fleet_server_host`) use `resource.ImportStatePassthroughID` and are broken for non-default spaces. The other four each have a hand-written `ImportState` method using `clients.CompositeIDFromStrFw` — but with slight behavioral variations (strict vs. lenient empty-segment validation, whether to also set `id`, whether to hardcode `"default"` on plain IDs).

All six resources embed `*resourcecore.Core` for Configure/Metadata wiring. That precedent motivates the same embedding pattern for import.

The `internal/fleet` package already owns `GetOperationalSpaceFromState` in `space_utils.go`. Adding `SpaceImporter` there keeps all fleet-space utilities co-located.

## Goals / Non-Goals

**Goals:**
- Fix the import bug in `fleet_output` and `fleet_server_host`
- Eliminate bespoke `ImportState` methods from all six resources via a shared embedding
- Establish a single canonical import behavior for all Fleet resources with `space_ids`
- Update OpenSpec requirements to match implementation across all four specs that are currently incorrect or incomplete

**Non-Goals:**
- Changing import behavior for resources without `space_ids` (non-Fleet resources)
- Adding import support to `fleet_integration` or `fleet_custom_integration` (they don't support import today)
- Changing how `space_ids` is written back during Read (unchanged)

## Decisions

### Decision 1: Embeddable struct over shared function

**Chosen:** `SpaceImporter` struct with promoted `ImportState` method, embedded in resource structs alongside `*resourcecore.Core`.

**Alternative:** A standalone function `fleet.HandleSpaceAwareImportState(ctx, req, resp, idField)` called from each resource's own `ImportState`.

**Rationale:** The embedding pattern is already established by `resourcecore.Core`. Embedding promotes the method directly, so resources that embed `*SpaceImporter` need no `ImportState` method at all — the interface is satisfied automatically. The standalone function approach still requires each resource to declare an `ImportState` method, keeping the boilerplate.

---

### Decision 2: `SpaceImporter` lives in `internal/fleet/`, not `internal/resourcecore/`

**Chosen:** `internal/fleet/space_importer.go`

**Alternative:** `internal/resourcecore/` alongside `Core`.

**Rationale:** `space_ids` as a schema field is specific to Fleet resources. Non-Fleet resources (Kibana, Elasticsearch) use different identity patterns. Putting this in `resourcecore` would give it false generality. The `internal/fleet` package already has `space_utils.go` — this is a natural neighbor.

---

### Decision 3: Variadic `idFields ...path.Path` constructor

**Chosen:** `NewSpaceImporter(fields ...path.Path)` — accepts one or more paths, all set to the resource ID on import.

**Alternative:** Single required `idField path.Path`.

**Rationale:** Needed for `agentdownloadsource`, which historically set both `id` and `source_id`. Even though `id` is the Terraform resource ID that Read will repopulate, the variadic form keeps the constructor uniform and avoids a special case. In practice all resources except `agentdownloadsource` use a single field.

---

### Decision 4: Plain import ID → `space_ids` NOT set (nil)

**Chosen:** When `CompositeIDFromStrFw` fails (no `/`), treat the whole string as the resource ID and leave `space_ids` unset.

**Alternative (current `agentdownloadsource` behavior):** Hardcode `space_ids = ["default"]`.

**Rationale:** `GetOperationalSpaceFromState` returns `""` for both nil and empty `space_ids`, and `""` routes to the default space in all Fleet API calls. The behaviors are observationally equivalent for resources in the default space. Not setting `space_ids` is more honest — "you didn't specify a space" vs. "I'm asserting default" — and avoids permanently writing `["default"]` into state for resources the user may later move or whose space scoping they haven't declared. This is a minor behavioral change for `agentdownloadsource` only; it does not affect any currently working import flow.

---

### Decision 5: `agentdownloadsource` drops `id` field assignment on import

**Chosen:** `NewSpaceImporter(path.Root("source_id"))` only; `id` is not set during `ImportState`.

**Rationale:** `id` is the Terraform resource ID (computed, managed by the framework). It is repopulated from `source_id` during the Read that immediately follows import. Setting it explicitly in `ImportState` was redundant. Dropping it removes a dual-path special case from `SpaceImporter`.

---

### Decision 6: Remove `integration_policy` strict empty-segment validation

**Chosen:** Align with the lenient behavior of `agent_policy` and `elastic_defend`: if `CompositeIDFromStrFw` parses successfully, use the result; if not, fall back to plain ID. Do not add explicit empty-segment checks.

**Rationale:** `CompositeIDFromStrFw` wraps `CompositeIDFromStr`, which splits on `/` and requires exactly two parts. An ID like `"space/"` or `"/resource"` produces two parts with an empty element — `CompositeIDFromStr` accepts these and returns `ClusterID=""` or `ResourceID=""`. The stricter validation in `integration_policy` was inconsistent with the other resources and adds complexity to `SpaceImporter`. Empty `ClusterID` routes to default space (harmless); empty `ResourceID` results in a 404 from the API, which produces a clear error. The extra validation is not worth the inconsistency.

## Risks / Trade-offs

- **`agentdownloadsource` plain-ID behavior change** — After migration, a plain import ID produces `space_ids = nil` instead of `space_ids = ["default"]`. Anyone who currently imports `agentdownloadsource` with a plain ID and relies on the post-import state having `space_ids = ["default"]` will see a difference in intermediate import state. The Read call converges correctly regardless, so this is not a regression in practice. → _Mitigation: document in changelog; update spec scenario._

- **`integration_policy` strict validation removal** — A degenerate ID like `"space/"` previously returned a clear diagnostic; after migration it returns a 404 from the API. The error is still clear, just at a different layer. → _Mitigation: acceptable trade-off for consistency._

- **Method promotion ambiguity** — If a future resource embeds both `*SpaceImporter` and another struct that also has `ImportState`, Go will require an explicit disambiguation method. This is a distant risk given no such struct exists. → _No mitigation needed now._

## Migration Plan

1. Add `internal/fleet/space_importer.go` with `SpaceImporter` and unit tests
2. Fix `fleet_output` and `fleet_server_host` (embed, wire, add acceptance tests) — bug fix scope
3. Migrate remaining four resources — cleanup scope
4. Update four delta specs
5. `make build && make lint` gate throughout; acceptance tests run in CI
