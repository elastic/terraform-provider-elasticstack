## Context

`ElasticsearchResource[T]`'s `writeFromPlan` currently calls the concrete callback and writes state directly from the returned model. Each callback is expected to call `readFunc` itself and handle the not-found case. This produces duplicated boilerplate in every resource's write function and creates an implicit contract that isn't enforced.

The read function is already registered with the envelope (for the standalone `Read` handler). This change moves the read-after-write step into `writeFromPlan` so the envelope owns the full create/update lifecycle.

## Goals / Non-Goals

**Goals:**
- `writeFromPlan` invokes `readFunc` after a successful concrete callback, using the callback's returned model as the prior state passed to `readFunc`
- Not-found after write produces a standard error diagnostic identifying the resource type
- Concrete callbacks no longer call `readFunc` or handle not-found
- No change to callback signatures

**Non-Goals:**
- Changing callback signatures to return only `diag.Diagnostics` (Path B — not viable because create-only fields such as API key secrets must travel from the callback to `readFunc` via the returned model)
- Refactoring `readFunc` implementations to compute composite IDs themselves
- Migrating resources not currently using the envelope

## Decisions

### Keep `(T, diag.Diagnostics)` callback signature

The concrete callback must still return the written model so the envelope can pass it to `readFunc` as prior state. This is necessary for resources where:
- The read function carries non-API fields through from the state parameter (composite ID, `ElasticsearchConnection`, create-only values)
- The create API response includes fields that will never be returned by a subsequent GET (e.g., API key secret values)

Changing to a `diag.Diagnostics`-only return (Path B) would make these resources impossible to implement correctly.

### Pass `writtenModel` (not plan model) to `readFunc`

`readFunc` takes a prior-state `T` to carry through fields the API does not return. Using the callback's returned model (`writtenModel`) rather than the original plan model ensures that:
- The composite ID is set correctly (callbacks that need it still call `client.ID()` and set it before returning)
- Create-only fields set by the callback survive into state
- `ElasticsearchConnection` is preserved

### Error message uses component and resource name from `ResourceBase`

`writeFromPlan` runs inside the `entitycore` package and has direct access to `r.component` and `r.resourceName`. The not-found error uses `fmt.Sprintf("%s_%s", r.component, r.resourceName)` to produce a type-qualified identifier (e.g. `elasticsearch_security_role`) without hardcoding the provider name.

## Risks / Trade-offs

**Callback must still set composite ID where readFunc carries it through** → Existing callbacks already do this; no behavior change required, but the implicit contract remains. Future callbacks must know to call `client.ID()` before returning. This is documented in the callback type's godoc.

**rolemapping readFunc computes its own ID via `client.ID()`** → This means `writeRoleMapping` no longer needs to call `client.ID()` at all, and can be simplified further. This is a minor bonus, not a risk.

## Open Questions

_None._
