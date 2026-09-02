## Why

`elasticstack_elasticsearch_security_role` fails with "Provider produced inconsistent result after apply" when a `dynamic "indices"` block appends a **new** element to an already-populated `indices` `SetNestedBlock` during `Update`, and the new element omits `allow_restricted_indices`. The failure only occurs on `Update` (not `Create`) and self-clears on the next `terraform apply`.

`allow_restricted_indices` is `Optional+Computed` with `boolplanmodifier.UseStateForUnknown()` on both `indices` and `remote_indices` entries. That plan modifier only replaces a planned-**unknown** value with the prior state value — it does nothing when the plan already carries a concrete `null`. On `Create`, Terraform Core marks the omitted computed attribute `Unknown` for every new element, so it reconciles cleanly against the `false` Elasticsearch returns. On `Update`, `indices` already has prior elements, and Core's best-effort Set-element correlation does not reliably promote the new element's omitted attribute to `Unknown` — it can survive as literal `null` in the final plan. Since Sets compare by whole-object equality, a planned element with `allow_restricted_indices: null` does not correlate with the actual element Elasticsearch returns (`allow_restricted_indices: false`), producing the "does not correlate with any element in actual" error.

## What Changes

- Replace the `Optional + Computed + boolplanmodifier.UseStateForUnknown()` definition of `allow_restricted_indices` with `Optional + Computed + Default: booldefault.StaticBool(false)` on both the `indices` and `remote_indices` nested blocks of the `elasticstack_elasticsearch_security_role` resource schema.
- `Default` fires whenever the **config** value is null, independent of whatever Core's Set-element correlation decided the intermediate planned value should be, so it is not vulnerable to the gap that defeats `UseStateForUnknown` for newly-appended Set elements.
- Matches the value Elasticsearch's Put Role API normalizes an omitted `allow_restricted_indices` to (`false`), so the change is a plan-time-only fix with no behavior change for the API payload or the read-back mapping.
- Add acceptance test coverage reproducing the issue's exact scenario (append a new `indices` element via a `dynamic` block on `Update`, omitting `allow_restricted_indices`) for both `indices` and `remote_indices`.
- Out of scope for this change: `field_security.except` and any other `Optional+Computed` nested-block attribute using `UseStateForUnknown()`, and migrating `indices`/`remote_indices` from `SetNestedBlock` to `ListNestedBlock`/`SetNestedAttribute`.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `elasticsearch-security-role`: replace REQ-023's `indices.allow_restricted_indices` and `remote_indices.allow_restricted_indices` "preserve prior state value when unknown" behavior with schema-level default semantics (`Default: booldefault.StaticBool(false)`), and update REQ-037 accordingly to describe the new plan modifier semantics for `remote_indices.allow_restricted_indices`.

## Impact

- `internal/elasticsearch/security/role/schema.go` — change `attrAllowRestrictedIndices` on both the `indices` and `remote_indices` `SetNestedBlock` definitions from `PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()}` to `Default: booldefault.StaticBool(false)`.
- `internal/elasticsearch/security/role/models_test.go` (and any other unit test asserting `allow_restricted_indices` stays `types.BoolNull()`) — update expectations, since the schema default now resolves the config-null case to `false` rather than leaving it null pre-apply.
- Acceptance tests (`internal/elasticsearch/security/role/`) — add a test reproducing the append-new-`indices`-element-on-update scenario from the issue, for both `indices` and `remote_indices`, asserting the second `apply` succeeds without a "does not correlate" error and without a required third apply.
- No API payload or read-mapping changes: `toAPIModel`/`fromAPIModel` in `models.go` already build/parse `allow_restricted_indices` correctly; only the plan-time default mechanism changes.
