## Context

The provider currently supports snapshot repositories (`elasticstack_elasticsearch_snapshot_repository`) and SLM policies (`elasticstack_elasticsearch_snapshot_lifecycle`), but provides no way to trigger the actual restore or ad-hoc create operations from Terraform. Users must fall back to external tooling or the REST API.

`terraform-plugin-framework v1.19.0` (already vendored at that version) ships `provider.ProviderWithActions`, `action.Action`, and `action.ActionWithConfigure`. Terraform Core 1.14+ is required on the client side; this is a new minimum that must be documented in each action's docs page.

`terraform-plugin-framework-timeouts v0.7.0` ships an `action/timeouts` package that provides action-specific timeout helpers (corrected from early research assumptions by `@tobio`). The `invoke` timeout from this package will be used instead of a hand-rolled schema attribute.

## Goals / Non-Goals

**Goals:**
- A `elasticstack_elasticsearch_snapshot_restore` action that calls `POST /_snapshot/{repo}/{snapshot}/_restore`.
- A `elasticstack_elasticsearch_snapshot_create` action that calls `POST /_snapshot/{repo}/{snapshot}`.
- Both support `wait_for_completion` (user-configurable) and an `invoke` timeout via `terraform-plugin-framework-timeouts`.
- `index_settings` on restore exposed as a JSON-encoded string attribute (matches other complex attributes in the provider; loses schema-level validation but is simpler).
- Acceptance tests bootstrap their own snapshot repository + snapshot (no shared CI repository available).
- Minimum Terraform version (1.14+) surfaced in each action's generated docs page.

**Non-Goals:**
- Trigger-based resource approach (Approach A) — superseded by actions.
- Cross-cluster restore.
- Searchable snapshot mount.
- SLM policy or snapshot repository changes.
- Idempotency guards for re-running restore over live indices — users are responsible; the action surfaces the ES error.

## Decisions

### D1: Terraform provider-defined actions (not trigger-based resource)

`provider.ProviderWithActions` is confirmed stable in v1.19.0. Actions are semantically correct for imperative one-shot operations: the plan output says "action will run", not "resource will be created", and there is no awkward no-op `Delete` handler. Both operations (restore and create) map cleanly onto this model.

### D2: `action/timeouts` from `terraform-plugin-framework-timeouts`

`@tobio` confirmed that `terraform-plugin-framework-timeouts v0.7.0` does include `action/timeouts` (see `https://github.com/hashicorp/terraform-plugin-framework-timeouts/tree/main/action/timeouts`). The implementation will use this package for the `timeouts` block, consistent with how resource timeouts are handled in the rest of the provider.

### D3: `index_settings` as JSON-encoded string

`index_settings` is a freeform `IndexSettings` object. Consistent with other complex freeform attributes in this provider (e.g., `metadata` in several resources), it will be exposed as a JSON-encoded string (`jsontypes.Normalized`). This trades schema-level validation for implementation simplicity. A follow-up can introduce structured attributes if there is demand.

### D4: Acceptance tests bootstrap full workflow

No shared snapshot repository is available in the CI environment. Each acceptance test will: (1) create a snapshot repository, (2) create a snapshot, (3) run the action under test. This requires a `TestAccAction*` naming convention and a pre-existing test infrastructure setup similar to SLM tests.

### D5: Package layout follows SLM pattern

New packages at `internal/elasticsearch/cluster/snapshot_restore/` and `internal/elasticsearch/cluster/snapshot_create/` mirror the SLM package at `internal/elasticsearch/cluster/slm/`. Client helpers at `internal/clients/elasticsearch/snapshot_restore.go` and `internal/clients/elasticsearch/snapshot_create.go` mirror `internal/clients/elasticsearch/snapshot_lifecycle.go`.

### D6: Provider registration

`provider/plugin_framework.go` adds:
1. `res.ActionData = factory` in `Configure()` alongside the existing `res.ResourceData`, `res.DataSourceData`, and `res.EphemeralResourceData` assignments.
2. A new `Actions(ctx context.Context) []func() action.Action` method on `*Provider`, making it satisfy `provider.ProviderWithActions`.
3. Both new action constructors are registered in the `Actions()` return slice.

### D7: Minimum Terraform version documentation

The 1.14+ requirement is non-obvious and affects managed Terraform Cloud users on older pinned versions. It will be documented in each action's generated docs page (the `description` field of the action's `Metadata` method, plus a note in the HCL usage example in the docs template). The provider README and compatibility table are out of scope for this change.

## Risks / Trade-offs

- **No action output**: `action.InvokeResponse` carries only `Diagnostics` and `SendProgress`; no typed result attributes. Downstream config expressions cannot reference restored index lists or shard counts. Progress events via `SendProgress` can give visibility during long-running restores, but result data is not composable into other resource expressions. This is a framework limitation, not a design choice.
- **Re-restore over live indices**: If the same restore action is re-applied against an existing index, the ES API returns an error (unless `rename_pattern` is used or indices are pre-deleted). This is intentional and the error is surfaced cleanly as a Terraform diagnostic. Users must manage idempotency themselves.
- **Large snapshot timeouts**: Restoring a large snapshot may take longer than the user configures. The `timeouts.invoke` attribute gives users control; the default should be generous (e.g., 20m).

## Open Questions

- **Timeouts block implementation for actions**: Confirmed resolved by `@tobio` — use `terraform-plugin-framework-timeouts v0.7.0` `action/timeouts` package.
- **Minimum Terraform version documentation**: Resolved — document in each action's generated docs page.
- **`index_settings` overrides on restore**: Resolved — JSON-encoded string.
- **Acceptance test infrastructure**: Resolved — tests bootstrap their own repository and snapshot.
