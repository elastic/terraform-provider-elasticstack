## Why

`elasticstack_elasticsearch_security_role` fails with "Provider produced inconsistent result after apply" when a `dynamic "indices"` block appends a **new** element to an already-populated `indices` `SetNestedBlock` during `Update`, and the new element omits `allow_restricted_indices`. The failure only occurs on `Update` (not `Create`) and self-clears on the next `terraform apply`.

`allow_restricted_indices` is `Optional+Computed` with `boolplanmodifier.UseStateForUnknown()` on both `indices` and `remote_indices` entries. That plan modifier only replaces a planned-**unknown** value with the prior state value — it does nothing when the plan already carries a concrete `null`. On `Create`, Terraform Core marks the omitted computed attribute `Unknown` for every new element, so it reconciles cleanly against the `false` Elasticsearch returns. On `Update`, `indices` already has prior elements, and Core's best-effort Set-element correlation does not reliably promote the new element's omitted attribute to `Unknown` — it can survive as literal `null` in the final plan. Since Sets compare by whole-object equality, a planned element with `allow_restricted_indices: null` does not correlate with the actual element Elasticsearch returns (`allow_restricted_indices: false`), producing the "does not correlate with any element in actual" error.

## What Changes

- Replace the `Optional + Computed + boolplanmodifier.UseStateForUnknown()` definition of `allow_restricted_indices` with `Optional + Computed + Default: booldefault.StaticBool(false)` on both the `indices` and `remote_indices` nested blocks of the `elasticstack_elasticsearch_security_role` resource schema.
- `Default` fires whenever the **config** value is null, independent of whatever Core's Set-element correlation decided the intermediate planned value should be, so it is not vulnerable to the gap that defeats `UseStateForUnknown` for newly-appended Set elements.
- Matches the value Elasticsearch's Put Role API normalizes an omitted `allow_restricted_indices` to (`false`). After the default resolves, `toAPIModel` will send `"allow_restricted_indices": false` instead of omitting the field; that is semantically identical for Elasticsearch and is required so the planned set element matches the read-back.
- Add a `resource.UnitTest` (mock Elasticsearch) that reproduces the issue's exact update-append sequence, plus live acceptance coverage for the same scenario on both `indices` and `remote_indices`.
- Out of scope for this change: `field_security.except` and any other `Optional+Computed` nested-block attribute using `UseStateForUnknown()`, migrating `indices`/`remote_indices` from `SetNestedBlock` to `ListNestedBlock`/`SetNestedAttribute`, and a state-schema version bump.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `elasticsearch-security-role`: narrow REQ-023–REQ-024 to `field_security.except` only; add REQ-039 for schema-level `Default: false` on `indices.allow_restricted_indices` and `remote_indices.allow_restricted_indices`; update REQ-037 to the same default semantics; update REQ-038 so an omitted config value is sent as `false` after default resolution. Update the capability Schema sketch to show `default false` on both blocks.

## Impact

- `internal/elasticsearch/security/role/schema.go` — change `attrAllowRestrictedIndices` on both the `indices` and `remote_indices` `SetNestedBlock` definitions from `PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()}` to `Default: booldefault.StaticBool(false)`. Do not bump `CurrentSchemaVersion`.
- `internal/elasticsearch/security/role/` tests — add a mock `resource.UnitTest` that creates one `indices` element (field omitted) then appends a second on update; add the equivalent live acceptance tests (remote steps gated at Elasticsearch 8.10). Do not rewrite hand-built `types.BoolNull()` fixtures in `models_test.go`; those are not plan-time values.
- Generated resource docs via `make docs-generate` — the schema `Default` will surface as the documented default. No wording change is required in `descriptions/allow_restricted_indices.md` unless it currently implies “unset means preserve prior state” (it does not today).
- Privilege reduction on the omitted-after-explicit-`true` path: a pre-existing set element whose state holds `allow_restricted_indices = true` but whose config now omits the field plans `false` and the next apply revokes restricted-index access. This is the REQ-039 contract; call it out in the changelog.
- `toAPIModel` / `fromAPIModel` mapping logic is unchanged. The Put Role payload for an omitted field changes from “omit key” to `"allow_restricted_indices": false` because the plan value is now known `false`. Read-back is already `false`.
