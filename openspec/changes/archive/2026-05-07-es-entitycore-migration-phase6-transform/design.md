## Context

The transform resource (`internal/elasticsearch/transform/transform.go`) is a 905-line Plugin SDK resource that manages Elasticsearch transforms. It uses the typed Elasticsearch client for API calls and has complex logic:

1. **Version-gated fields** — `destination.aliases` (≥8.8), `deduce_mappings` (≥8.1), `num_failure_retries` (≥8.4), `unattended` (≥8.5) are silently omitted when the server version is too low.
2. **`pivot`/`latest` mutual exclusivity** — Exactly one must be set. Both are `ForceNew`.
3. **Start/stop lifecycle** — `enabled` controls whether Start Transform or Stop Transform is called after Put/Update Transform.
4. **Defer validation** — Passed as a query parameter to Put and Update Transform.
5. **Complex nested blocks** — `source`, `destination`, `retention_policy`, `sync`, and a flat list of settings attributes.
6. **JSON fields** — `source.query`, `source.runtime_mappings`, `pivot`, `latest`, `metadata` are JSON strings with diff suppression.

The entitycore resource envelope already centralizes the standard PF resource prelude. No envelope changes are required.

## Goals / Non-Goals

**Goals:**

- Rewrite the resource from SDK to PF while preserving exact schema shape and behavior.
- Migrate to the entitycore envelope.
- Preserve version gating, start/stop lifecycle, JSON handling, and import.

**Non-Goals:**

- Changing the Terraform schema or acceptance tests.
- Changing the underlying typed-client helpers in `internal/clients/elasticsearch`.
- Centralizing transform-specific patterns into the envelope.

## Decisions

### D1. Real create callback; override Update

**Choice:** Pass a real create callback, real read callback, real delete callback, and a placeholder update callback. Define `Update` on the concrete type.

**Rationale:**
- Create only needs the plan model. It calls Put Transform and optionally Start Transform. This fits the callback contract.
- Update needs to compare old `enabled` with new `enabled` to decide whether to call Start or Stop Transform after Update Transform. The envelope's update callback only receives the planned model, not the prior state.

### D2. Schema replicates SDK shape in PF

**Choice:** Express every SDK attribute as a PF equivalent using the same validation rules, defaults, and plan modifiers (`UseStateForUnknown` where appropriate, `RequiresReplace` for force-new attributes).

**Rationale:** Exact external parity. The schema is large but mechanical.

### D3. Version gating preserved via model builders

**Choice:** Keep the `isSettingAllowed` helper logic inside the model-to-API conversion functions, checking `client.ServerVersion` before including gated fields in the request.

**Rationale:** This logic is resource-specific and does not belong in the envelope.

### D4. `pivot`/`latest` use PF `ExactlyOneOf` validator

**Choice:** Use `stringvalidator.ExactlyOneOf(path.Expressions{...})` in the PF schema to enforce that exactly one of `pivot` or `latest` is set.

**Rationale:** PF provides native `ExactlyOneOf` validation. The SDK used `schema.ExactlyOneOf`.

### D5. JSON fields use `jsontypes.Normalized` with diff suppression

**Choice:** Use `jsontypes.NormalizedType` for all JSON string attributes, preserving JSON-normalized diff suppression.

**Rationale:** Matches the SDK's `DiffSuppressFunc: tfsdkutils.DiffJSONSuppress`.

### D6. `enabled` derived from transform stats on read

**Choice:** The read callback calls both Get Transform and Get Transform Stats. It sets `enabled = true` when the stats state is `"started"` or `"indexing"`.

**Rationale:** Preserves current behavior.

### D7. ImportState preserved

**Choice:** Implement `ImportState` as passthrough on `id`.

**Rationale:** The SDK resource already supported import.

## Risks / Trade-offs

- **Risk:** SDK→PF rewrite surfaces subtle differences in set/list/map handling. **Mitigation:** Acceptance tests cover the full create/update/read/delete cycle.
- **Risk:** JSON diff suppression behavior may drift between SDK and PF. **Mitigation:** `jsontypes.Normalized` handles this natively.
- **Risk:** Version-gated settings omitted incorrectly. **Mitigation:** Unit tests for `isSettingAllowed` equivalents and acceptance tests on multi-version clusters.
- **Risk:** Start/stop lifecycle timing. **Mitigation:** Acceptance tests with `enabled = true` and `enabled = false` verify the lifecycle.

## Migration Plan

1. Define `tfModel` with all PF types matching the SDK schema.
2. Add `GetID()`, `GetResourceID()`, `GetElasticsearchConnection()`.
3. Write `getSchema() schema.Schema` factory with all validators and plan modifiers.
4. Write model-to-API conversion helpers (`toAPIModel`) preserving version gating.
5. Write API-to-model conversion helpers (`fromAPIModel`) for read.
6. Implement `readTransform`, `createTransform`, `deleteTransform` callbacks.
7. Implement `Update` override with enabled-change detection and start/stop.
8. Wire into provider registrar.
9. Run `make build`, `make check-lint`, `make check-openspec`, and acceptance tests.

## Open Questions

None.
