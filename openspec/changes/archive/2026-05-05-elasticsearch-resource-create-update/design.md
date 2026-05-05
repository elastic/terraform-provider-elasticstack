## Context

`NewElasticsearchResource` currently owns common Elasticsearch resource behavior for Metadata, Schema, Read, Delete, and Configure. Concrete resources still implement `Create` and `Update`, and several migrated resources repeat the same thin Plugin Framework wrapper around a shared upsert helper.

The duplicate wrappers are visible in `security_role`, `security_role_mapping`, `security_system_user`, and `cluster_script`. Their create and update operations share a common shape: decode the plan, resolve an Elasticsearch client from `elasticsearch_connection`, perform an API write, and persist the resulting model to state.

## Goals / Non-Goals

**Goals:**

- Make `NewElasticsearchResource` a complete `resource.Resource` for Elasticsearch-backed resources that fit the standard lifecycle.
- Require separate create and update callbacks so resources can pass the same function when behavior is identical or distinct functions when behavior diverges.
- Keep state persistence centralized in the envelope by having callbacks return the final model.
- Pass the resource ID to create and update callbacks, matching the existing read and delete callback style while allowing create plans whose computed `id` is not yet known.
- Migrate only the four duplicated resources identified in the issue.

**Non-Goals:**

- Do not introduce `NewUpsertableElasticsearchResource` or a generic Create/Update embeddable.
- Do not migrate unrelated Elasticsearch, Kibana, Fleet, or APM resources opportunistically.
- Do not change public Terraform schemas, state formats, IDs, or API behavior.
- Do not make create or update callbacks optional; resources that do not fit the envelope should continue to use `ResourceBase` directly **or** embed `ElasticsearchResource` while supplying non-nil placeholder callbacks from `PlaceholderElasticsearchWriteCallbacks` and overriding `Create` and `Update` on the concrete type when the envelope prelude cannot supply required inputs (for example write-only attributes read from `Config`). The `security_user` resource follows this pattern.

## Decisions

### Extend the Elasticsearch envelope instead of adding a second embeddable

`NewElasticsearchResource` already owns the common Elasticsearch request prelude for Read and Delete: model decoding, composite ID parsing, scoped client resolution, and state mutation. Extending the same envelope to Create and Update keeps the abstraction boundary coherent.

Alternative considered: add an `UpsertableResource` embeddable that implements Create and Update from a `func(context.Context, tfsdk.Plan, *tfsdk.State) diag.Diagnostics`. This would remove the thin wrapper methods but leave client resolution and state mutation inside concrete resources, and receiver-bound initialization would be awkward because existing helpers call `r.Client()`.

### Require two lifecycle callbacks

The constructor should accept both create and update callbacks. The duplicated resources can pass the same function twice, but the API should not assume create and update are always identical.

Alternative considered: accept a single upsert callback. This would fit the four initial resources but would make the envelope less useful for resources whose create and update operations diverge later.

### Return the final model from callbacks

Create and update callbacks should return `(T, diag.Diagnostics)` instead of mutating `*tfsdk.State` directly. The envelope should append diagnostics and call `resp.State.Set(ctx, &resultModel)` only after the callback succeeds.

This mirrors the existing Read contract, where concrete logic returns a model and the envelope owns persistence. It also makes callback tests less coupled to Plugin Framework state plumbing.

### Derive resource ID from the planned model for create and update

The create and update prelude should decode the planned model, resolve the scoped Elasticsearch client, derive the write resource ID from the planned model, and invoke the callback with `(ctx, client, resourceID, model)`.

Create cannot generally parse `model.GetID()`, because `id` is computed for these resources and may be unknown in the plan. The envelope should therefore extend `ElasticsearchResourceModel` with a plan-safe resource identity accessor, for example `GetResourceID() types.String`, implemented by each model from its natural write identity field (`name`, `username`, or `script_id`). Read and Delete should continue to parse the composite state ID via `GetID()`.

This keeps callback signatures consistent with Read and Delete without forcing create callbacks to recover the same identity from the model themselves. Callbacks remain responsible for assigning the final composite `ID` on the returned model after successful API writes.

## Risks / Trade-offs

- Constructor signature churn → Update all current `NewElasticsearchResource` call sites in the same implementation change so compile errors surface missed migrations.
- Existing envelope models do not expose a separate write identity accessor → Add the accessor to the model constraint and update current envelope model types in the same change.
- Larger refactor than thin wrapper extraction → Keep migration scoped to the four issue examples and preserve existing acceptance-visible behavior.
- Existing spec currently says the envelope does not implement Create or Update → Update the delta spec to make the contract change explicit.
