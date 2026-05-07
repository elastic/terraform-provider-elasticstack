## Context

Both `ml_job_state` and `ml_datafeed_state` are state-transition resources. They do not create or delete underlying Elasticsearch entities; they drive transitions (open/close for jobs, start/stop for datafeeds) and are removed from Terraform state on destroy. Their current implementations embed `*entitycore.ResourceBase` and have:

- `Create` and `Update` that extract framework timeouts, create a sub-context with that timeout, and delegate to a shared `update()` helper.
- `Delete` that is a no-op for job_state (logs and returns) and that stops the datafeed if running for datafeed_state.
- `Read` that queries stats APIs and defaults computed attributes.

The envelope's standard Create/Update/Delete callback contract does not fit these resources cleanly because:
- The write callbacks would need to extract `timeouts` from the model and manage their own sub-contexts.
- The job_state Delete is intentionally a no-op.
- The datafeed_state Delete conditionally calls Stop Datafeed.

## Goals / Non-Goals

**Goals:**

- Migrate both resources to the entitycore envelope.
- Keep all state-transition behavior exactly as-is.
- Preserve ImportState support.

**Non-Goals:**

- Changing the transition logic or timeout handling.
- Changing schemas or acceptance tests.

## Decisions

### D1. Override Create, Update, and Delete; use real read callbacks

**Choice:** Both resources pass placeholder write callbacks for create/update and define Create, Update, and Delete methods on the concrete types. Real read callbacks are passed to the envelope. Non-nil delete callbacks are still provided for envelope construction; concrete Delete methods preserve the existing no-op and conditional-stop behavior.

**Rationale:** The envelope's write callback contract assumes a standard create/update/delete API call. These resources have custom timeout handling, conditional API calls, and a no-op delete. Overriding keeps all existing logic intact.

### D2. Read callbacks default computed attributes during import

**Choice:** The read callbacks set default values for computed attributes (`force`, `job_timeout` for job_state; `force`, `datafeed_timeout`, `start`, `end` for datafeed_state) when they are null, matching the current Read behavior.

**Rationale:** This is required for import to produce a complete state.

### D3. Keep ImportState on concrete types

**Choice:** Both resources implement `ResourceWithImportState` themselves.

**Rationale:** The envelope does not implement ImportState. Job_state uses passthrough on `id`. Datafeed_state uses passthrough on `datafeed_id`.

### D4. Schema factories strip connection block

**Choice:** Both schema factories return `schema.Schema` without `elasticsearch_connection`.

**Rationale:** Standard envelope convention.

## Risks / Trade-offs

- **Risk:** Minimal — these are thin resources and the override pattern is well-proven from other migrations. The envelope removes ~30 lines of duplicated Read/Schema code per resource.
- **Trade-off:** Placeholder callbacks add a tiny bit of boilerplate. Acceptable for the clarity of keeping transition logic on the concrete type.

## Migration Plan

1. For both `jobstate` and `datafeed_state`:
   - Add getters to the model.
   - Convert schema to a factory.
   - Extract read logic into a package-level read callback.
   - Replace `*ResourceBase` with `*entitycore.ElasticsearchResource[T]`.
   - Use placeholder write callbacks.
   - Keep `Create`, `Update`, `Delete`, and `ImportState` on the concrete type.
2. Run `make build`, `make check-lint`, `make check-openspec`, and acceptance tests.

## Open Questions

None.
