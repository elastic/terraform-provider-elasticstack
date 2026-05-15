## Context

The current Elasticsearch entitycore resource envelope centralizes Schema, Configure, Metadata, Read, Delete, and a narrow Create/Update flow for CRUD-style Plugin Framework resources. Its current API assumes:

- create/update callbacks only need the planned model and a write identity
- read-after-write should always use that same write identity
- version requirement enforcement is Kibana-only
- post-read side effects are not part of the envelope contract
- the constructor can remain positional because the configuration surface is small

That contract was enough for early migrations, but several Elasticsearch resources now override Create, Update, or Read for reasons that are not unique lifecycles so much as missing callback context or missing lifecycle extension points.

The goal of this change is to evolve the existing Elasticsearch envelope into a richer CRUD-oriented envelope without trying to absorb fundamentally different state-transition resources.

## Goals / Non-Goals

**Goals:**

- Replace the positional Elasticsearch envelope constructor with a named options form.
- Remove the redundant `component` parameter from the Elasticsearch envelope constructor.
- Add structured create/update callback request types that expose the correct Terraform inputs.
- Add support for model-declared version requirements in the Elasticsearch envelope using the same semantics as Kibana.
- Add support for a stable read identity distinct from the write identity via an optional model interface.
- Add an optional post-read hook so standard read flows can persist private state or perform other read-side side effects without overriding `Read`.
- Prove the new envelope API on five existing Elasticsearch resources.

**Non-Goals:**

- Creating a separate Elasticsearch transition/state-resource envelope.
- Reshaping the Elasticsearch read callback signature.
- Generalizing read-after-write beyond the standard envelope pattern.
- Migrating every remaining Elasticsearch override in this change.
- Changing external Terraform schemas or behavior except where envelope internals become less override-driven.

## Decisions

### Decision 1: Replace the positional constructor with `NewElasticsearchResource(name, opts)`

**Choice:** The Elasticsearch envelope constructor will become:

- `NewElasticsearchResource[T](name string, opts ElasticsearchResourceOptions[T])`

where `ElasticsearchResourceOptions[T]` holds the schema factory and lifecycle callbacks.

**Rationale:**
- The current positional signature is already long and difficult to extend.
- The Elasticsearch envelope always produces Elasticsearch namespaced resources, so repeating `entitycore.ComponentElasticsearch` is redundant.
- The envelope now needs optional hooks and richer callback types; an options struct makes this tractable and readable.

**Rejected alternative:** fluent/builder chaining (e.g. `NewElasticsearchResource(...).WithCreate(...)`). Rejected because an options struct is simpler, more idiomatic for this repository, and easier to validate/document as a single configuration object.

### Decision 1a: Kibana envelope constructor alignment is out of scope

**Choice:** This change does not migrate the Kibana resource envelope to the same options-struct constructor pattern.

**Rationale:**
- The Elasticsearch envelope is the current pressure point and can be evolved independently.
- Constructor parity between Elasticsearch and Kibana is desirable, but not required to unblock the richer Elasticsearch CRUD envelope.
- If the Elasticsearch options-struct pattern proves successful, Kibana alignment can be proposed as a follow-up refactor.

### Decision 2: Keep one richer CRUD-oriented Elasticsearch envelope

**Choice:** Evolve the existing Elasticsearch envelope rather than adding a second CRUD envelope at this time.

**Rationale:**
- The immediate pressure comes from CRUD-style resources needing more context, not from a fundamentally different CRUD lifecycle.
- A second CRUD envelope would introduce premature choice for authors without a crisp enough boundary.
- State-transition resources remain out of scope and may justify a separate abstraction later.

### Decision 3: Introduce structured create/update callback request types

**Choice:** Replace the narrow callback parameters with request structs:

- `ElasticsearchCreateRequest[T]` with `Plan`, `Config`, `WriteID`
- `ElasticsearchUpdateRequest[T]` with `Plan`, `Prior`, `Config`, `WriteID`

The create and update callbacks receive `(ctx, client, req)` and return `ElasticsearchWriteResult[T]`.

**Rationale:**
- `Prior` is required by multiple update flows (`ml_filter`, `cluster_settings`, `ml_anomaly_detection_job`).
- `Config` is required by `security_user` and is generally the correct source for write-only and plan-divergent config semantics.
- Request structs avoid future positional-signature churn and keep the envelope readable.

### Decision 4: Keep the read callback signature unchanged

**Choice:** Continue to use the current Elasticsearch read callback signature:

- `func(context.Context, *clients.ElasticsearchScopedClient, string, T) (T, bool, diag.Diagnostics)`

**Rationale:**
- Current read callback shape is already sufficient for the standard read pipeline.
- This keeps the first proposal smaller and reduces migration churn.
- The main missing read capability is not callback input shape but post-read hook support and stable read identity resolution.

### Decision 5: Generalize stable read identity with `WithReadResourceID`

**Choice:** Add an optional model interface:

- `WithReadResourceID`

This interface is general to all read flows, not just read-after-write.

**Rationale:**
- Read-after-write is an implementation detail. If a resource has a canonical refresh identity, that identity should be available to both ordinary refresh and post-write refresh.
- `elasticstack_elasticsearch_index` already demonstrates the need for a stable refresh identity distinct from the configured/write identity.

**Read identity resolution rule:**

The envelope SHALL use one shared read identity resolution path for ordinary `Read` and read-after-write:

```text
1. If the model implements WithReadResourceID and returns a non-empty value, use it.
2. Otherwise, for ordinary Read/Delete, use the composite-id resource segment parsed from state.ID.
3. Otherwise, for read-after-write, use WriteID.
```

This keeps a stable, model-driven read identity central without introducing an explicit per-write read override.

### Decision 6: Do not add an explicit `ReadResourceID` field to write results

**Choice:** `ElasticsearchWriteResult[T]` will carry only the returned model reference, not a separate explicit read-id override.

**Rationale:**
- In Terraform provider semantics, the refresh identity needed by `readFunc` should normally be a stable property of the resource model/state rather than an ephemeral write artifact.
- No current Elasticsearch resource clearly requires an operation-only read identity that cannot be represented as model state.
- Avoiding an explicit override keeps the API smaller and identity semantics clearer.

### Decision 7: Add `PostRead` hook to the envelope

**Choice:** Add an optional `PostRead` callback to `ElasticsearchResourceOptions[T]` that runs after a successful state-setting read flow.

**Rationale:**
- `elasticstack_elasticsearch_security_api_key` and `elasticstack_elasticsearch_index` both need standard read behavior plus private-state side effects.
- This is a clean extension point that removes a common reason to override `Read`.

**Semantics:**
- The hook runs after successful read result handling and state persistence.
- The hook does not run when the resource is not found, when read returns diagnostics, or when state set fails.

### Decision 8: Extend the Elasticsearch envelope to honor `WithVersionRequirements`

**Choice:** Reuse the same optional `WithVersionRequirements` interface and enforce it with the same lifecycle semantics as Kibana.

**Rationale:**
- The same model-driven, client-scoped version guard pattern applies equally to Elasticsearch resources.
- Several Elasticsearch resources currently implement version checks manually.
- Shared behavior across envelopes reduces inconsistency.

**Semantics:**
- Create: evaluate requirements against the planned model after client resolution and before invoking the create callback.
- Update: evaluate requirements against the planned model after client resolution and before invoking the update callback.
- Read: evaluate requirements against the decoded state model after client resolution and before invoking the read callback.

### Decision 9: Rename `DataSourceVersionRequirement` to `VersionRequirement`

**Choice:** Replace the current requirement type name with a resource/data-source neutral shared name.

**Rationale:**
- The type is used beyond data sources and there is nothing data-source-specific about it.
- The rename matches the broadened, cross-envelope usage.

## Proposed API Shape

### Optional model interfaces

```text
ElasticsearchResourceModel
- GetID() types.String
- GetResourceID() types.String
- GetElasticsearchConnection() types.List

WithReadResourceID (optional)
- GetReadResourceID() string

WithVersionRequirements (optional)
- GetVersionRequirements() ([]VersionRequirement, diag.Diagnostics)
```

### Callback request/result types

```text
ElasticsearchCreateRequest[T]
- Plan T
- Config tfsdk.Config
- WriteID string

ElasticsearchUpdateRequest[T]
- Plan T
- Prior T
- Config tfsdk.Config
- WriteID string

ElasticsearchWriteResult[T]
- Model T
```

### Callback types

```text
Read(ctx, client, resourceID, stateModel) -> (model, found, diags)
Delete(ctx, client, resourceID, stateModel) -> diags
Create(ctx, client, createReq) -> (writeResult, diags)
Update(ctx, client, updateReq) -> (writeResult, diags)
PostRead(ctx, client, model, privateState) -> diags   // optional
```

### Constructor configuration

```text
NewElasticsearchResource[T](name string, opts ElasticsearchResourceOptions[T])

ElasticsearchResourceOptions[T]
- Schema   func(context.Context) rschema.Schema
- Read     ElasticsearchReadFunc[T]
- Delete   ElasticsearchDeleteFunc[T]
- Create   ElasticsearchCreateFunc[T]
- Update   ElasticsearchUpdateFunc[T]
- PostRead ElasticsearchPostReadFunc[T] // optional
```

## Resource Mapping to the New Envelope

### Strong fit / target migrations in this change

#### `elasticstack_elasticsearch_ml_filter`
- Current override reason: update performs a remote GET to compute item add/remove diffs and also needs prior state for description equality checks.
- New API fit: `UpdateRequest.Prior` removes the need for the concrete `Update` override's prior-state access, while the callback may still perform the remote GET to preserve current stale-state behavior.
- POC role: smallest prior-aware update migration that still preserves a remote reconciliation step.

#### `elasticstack_elasticsearch_cluster_settings`
- Current override reason: update needs prior state to null out removed settings.
- New API fit: `UpdateRequest.Prior` removes the need for the concrete `Update` override.
- POC role: prior-aware update with map nulling/merge semantics.

#### `elasticstack_elasticsearch_ml_anomaly_detection_job`
- Current override reason: update body is built from plan vs prior state.
- New API fit: `UpdateRequest.Prior` removes the need for the concrete `Update` override.
- POC role: more complex partial-update builder using prior state.

#### `elasticstack_elasticsearch_security_user`
- Current override reason: create/update need raw config for write-only password handling and prior state for password change detection.
- New API fit: `CreateRequest.Config`, `UpdateRequest.Config`, and `UpdateRequest.Prior` remove the need for concrete create/update overrides.
- POC role: validates config-aware callbacks.

#### `elasticstack_elasticsearch_security_api_key`
- Current override reason: read path persists cluster-version private state after a successful refresh.
- New API fit: `PostRead` removes the need for the concrete `Read` override.
- POC role: validates post-read side effects and private-state integration.

### Not targeted in this change

#### `elasticstack_elasticsearch_index`
- Benefits from `WithReadResourceID` and `PostRead`, but still has create adoption and identity complexity.
- Valuable future migration target, but not a first-wave proof-of-concept in this proposal.

#### `elasticstack_elasticsearch_transform`
- Benefits from prior-aware update callbacks, but still carries version-aware conversion and lifecycle-specific orchestration.
- Out of scope for first-wave proof-of-concept migration.

#### `elasticstack_elasticsearch_ml_job_state` and `elasticstack_elasticsearch_ml_datafeed_state`
- These are state-transition resources with timeout/wait orchestration and remain out of scope for this CRUD-envelope evolution.

## Risks / Trade-offs

- **Risk:** The Elasticsearch envelope becomes broader and more capable.  
  **Mitigation:** Keep the scope focused on CRUD-style resources and avoid adding skip-read or transition semantics in this change.

- **Risk:** Constructor migration touches many resources even if their behavior does not otherwise change.  
  **Mitigation:** Limit semantic resource migrations to the five proof-of-concept targets, and treat the broader constructor update as mechanical.

- **Risk:** Introducing `WithReadResourceID` may create ambiguity if a model's returned read identity diverges from its stored composite ID resource segment.  
  **Mitigation:** Document a single envelope-wide read identity resolution rule and add targeted tests covering both ordinary read and read-after-write.

- **Risk:** Version requirement enforcement on Elasticsearch resources could surface earlier errors than today for resources that currently defer checks deeper into create/update bodies.  
  **Mitigation:** Apply the same semantics as Kibana, migrate resource checks intentionally, and keep version-aware conversion logic in callbacks where appropriate.

- **Risk:** `ml_filter` item reconciliation semantics could regress if the migration switches from remote GET diffing to pure prior-state diffing.  
  **Mitigation:** Preserve the existing remote GET inside the migrated update callback and use `UpdateRequest.Prior` only for description equality and envelope-owned prelude removal.

- **Risk:** `security_api_key` may perform two `ServerVersion` calls during refresh: one inside the read callback and one inside the post-read hook that persists cluster-version private state.  
  **Mitigation:** Accept the duplicate call in the proof-of-concept migration, document it, and optimize later only if it proves materially costly.

## Migration Plan

1. Rename `DataSourceVersionRequirement` to `VersionRequirement` and update shared version requirement code/tests plus Kibana model implementations/spec references that use the old type name.
2. Introduce `ElasticsearchResourceOptions[T]`, request/result structs, `WithReadResourceID`, and `PostRead` support in the envelope.
3. Generalize `enforceVersionRequirements` (or add an Elasticsearch-specific variant) so Elasticsearch envelopes can enforce the same `WithVersionRequirements` contract as Kibana.
4. Update Elasticsearch envelope tests to cover the new constructor shape, request structs, version requirements, read identity resolution, and post-read hook behavior.
5. Mechanically migrate existing Elasticsearch envelope call sites to `NewElasticsearchResource(name, opts)`; prefer a scripted refactor for the broad constructor change and validate with `make build`.
6. Migrate the five proof-of-concept resources:
   - `ml_filter`
   - `cluster_settings`
   - `ml_anomaly_detection_job`
   - `security_user`
   - `security_api_key` (read/post-read path)
7. Update OpenSpec requirements for the envelope and migrated resources.
8. Verify build, lint, OpenSpec validation, and focused tests/acceptance tests for affected resources.
   - `ml_filter`
   - `cluster_settings`
   - `ml_anomaly_detection_job`
   - `security_user`
   - `security_api_key` (read/post-read path)
6. Update OpenSpec requirements for the envelope and migrated resources.
7. Verify build, lint, OpenSpec validation, and focused tests/acceptance tests for affected resources.

## Open Questions

- Should the Kibana resource envelope migrate to the same options-struct constructor pattern in a follow-up to keep constructor parity if the Elasticsearch pattern proves successful?
- Beyond the five proof-of-concept migrations in this change, should `elasticstack_elasticsearch_index` be the first follow-up target to validate `WithReadResourceID` on a more identity-complex resource?
