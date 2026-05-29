# Write-only secret drift detection

This document is the canonical reference for any provider resource that exposes write-only secret material (passwords, API tokens, connector secrets). Use the shared helper at [`internal/utils/writeonlyhash`](../../internal/utils/writeonlyhash/) ([package Godoc](https://pkg.go.dev/github.com/elastic/terraform-provider-elasticstack/internal/utils/writeonlyhash)) to store bcrypt hashes of applied values in resource private state and detect silent in-config edits at plan time without revealing the secret.

## Why hash-in-private-state over `_wo_version` companions

Some resources in this provider (for example `elasticstack_kibana_action_connector` with `secrets_wo` / `secrets_wo_version`) use a companion version attribute: the practitioner bumps the version when the secret rotates so Terraform schedules an update. That pattern works but depends on user discipline. If someone edits the secret in configuration without bumping the version, Terraform may not plan an update and the new value never reaches the API.

Hash-in-private-state compares the configured write-only value against the hash of the last successfully applied value. A changed secret in configuration is detected automatically during `ModifyPlan`, with a warning diagnostic naming the attribute path (never the value). The one-time cost of wiring `ModifyPlan`, Create/Update persistence, and removal cleanup is amortised across every secret-bearing resource via the shared helper.

## Threat model

Terraform state files (including remote backends) can leak through misconfiguration, compromised credentials, or backup exposure. Storing plaintext secrets in state is avoided by marking attributes `WriteOnly` and `Sensitive`, but practitioners still type secrets into configuration.

If the provider stored only a fast hash (for example SHA-256) of those secrets in private state, an attacker with a state dump could brute-force low-entropy passwords offline. The helper therefore uses **bcrypt** with default cost 10 (~100ms per hash at plan/apply time), which is negligible for typical applies but materially slows offline guessing.

Each Terraform resource type constructs its own `Hasher` via `writeonlyhash.New("<resource_type_name>")`, which derives a per-resource-type salt. The same plaintext secret therefore produces different hashes under different resource types, reducing cross-resource rainbow-table value. This matches the precedent set by the HashiCorp [`random_password`](https://registry.terraform.io/providers/hashicorp/random/latest/docs/resources/password#bcrypt_hash) resource, which exposes a computed `bcrypt_hash` for the generated password.

## `ModifyPlan` contract

Implement `resource.ResourceWithModifyPlan` on the resource (or equivalent envelope hook). **Read must not read or write private-state hashes**; the stored hash always represents the last value successfully sent to the API.

### Write-only attribute present in config

For each write-only secret attribute set in `req.Config`:

1. Read the string value from config (write-only values are available in config during plan, not in state after apply).
2. `hash, err := hasher.Compute(value)` — propagate errors as diagnostics; error strings from the helper never include the input value.
3. `storedHash, diags := req.Private.GetKey(ctx, hasher.PrivateStateKey(attributePath))` where `attributePath` is the Terraform attribute path (see [Private-state key convention](#private-state-key-convention)).
4. If `storedHash` is absent or empty: **first apply or post-import** — no drift signal; do not emit a warning.
5. If `hasher.Matches(value, storedHash)`: no drift.
6. If the stored hash exists and does **not** match: **mark the resource for update** (see worked example) and emit a **warning** diagnostic, for example: `Detected a change to write-only attribute api_token; the resource will be updated.` The diagnostic must name the attribute path only; never include the secret value.

### Write-only attribute removed from config

When a write-only attribute (or a map element containing one) is removed from the new configuration, clear the corresponding private-state entry: `resp.Private.SetKey(ctx, hasher.PrivateStateKey(attributePath), nil)`.

### After successful Create or Update

Persist a hash for every write-only secret value that was applied to the API:

```go
hash, err := hasher.Compute(appliedValue)
// ...
resp.Private.SetKey(ctx, hasher.PrivateStateKey("api_token"), hash)
```

### Read

Do not call `GetKey` / `SetKey` on private state for `secret_hash:*` keys during Read.

## Private-state key convention

Keys are `secret_hash:<attribute_path>`, where `<attribute_path>` matches the Terraform attribute path string passed to `Hasher.PrivateStateKey`. The helper prepends the `secret_hash:` prefix; callers supply only the path.

Examples:

| Attribute | `PrivateStateKey` argument | Stored key |
| --- | --- | --- |
| Flat attribute | `"aws.external_id"` | `secret_hash:aws.external_id` |
| Map element secret | `` `configuration_values["password"].secret_value` `` | `` secret_hash:configuration_values["password"].secret_value `` |

Use bracket notation with quoted map keys in paths for nested map elements, consistent with Terraform's attribute path syntax.

## Post-import behaviour

After `terraform import`, no `secret_hash:*` entries exist in private state. The first **refresh** (`terraform plan` with no config change) therefore produces **no drift signal** for write-only secrets, even if the live secret differs from what the practitioner will configure. The first **apply** that sets a write-only value in config computes and stores the hash, establishing the baseline. This mirrors [`random_password.bcrypt_hash`](https://registry.terraform.io/providers/hashicorp/random/latest/docs/resources/password#bcrypt_hash).

Recommend documenting this in user-facing resource description text for each adopter so practitioners know to run `terraform apply` after import when managing secrets.

## Worked example

The following fictional resource `elasticstack_example_thing` has one write-only attribute `api_token`. The [content connector resource spec](../../openspec/changes/elasticsearch-content-connector/specs/elasticsearch-content-connector/spec.md) (REQ-011) shows the same helper calls on `configuration_values["<key>"].secret_value` with `writeonlyhash.New("elasticsearch_connector")`.

### Schema

```go
"api_token": schema.StringAttribute{
    Description: "API token for the example integration. Write-only; not stored in state.",
    Optional:    true,
    Sensitive:   true,
    WriteOnly:   true,
},
```

### Hasher construction

Construct one `Hasher` per resource type at package scope:

```go
var apiTokenHasher = writeonlyhash.New("elasticstack_example_thing")
```

Use a stable, unique string per resource type. Do not share one `Hasher` across different resource types.

### `ModifyPlan`

```go
func (r *exampleThingResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
    const attributePath = "api_token"
    key := apiTokenHasher.PrivateStateKey(attributePath)

    var config, plan, state exampleThingModel
    resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
    resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
    if req.State.Raw != nil {
        resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
    }
    if resp.Diagnostics.HasError() {
        return
    }

    // Attribute removed from config — clear stored hash.
    if config.APIToken.IsNull() {
        if !state.APIToken.IsNull() || !plan.APIToken.IsNull() {
            resp.Diagnostics.Append(resp.Private.SetKey(ctx, key, nil)...)
        }
        return
    }
    if !typeutils.IsKnown(config.APIToken) {
        return
    }

    value := config.APIToken.ValueString()
    storedHash, diags := req.Private.GetKey(ctx, key)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }

    if len(storedHash) > 0 && !apiTokenHasher.Matches(value, storedHash) {
        resp.Diagnostics.AddWarning(
            "Write-only attribute changed",
            fmt.Sprintf("Detected a change to write-only attribute %s; the resource will be updated.", attributePath),
        )
        // Mark for update: ensure the plan carries the configured write-only value so
        // Terraform schedules an apply (state does not retain write-only values).
        resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("api_token"), config.APIToken)...)
    }
}
```

### After Create / Update

After the API accepts the new token:

```go
hash, err := apiTokenHasher.Compute(appliedToken)
if err != nil {
    resp.Diagnostics.AddError("Failed to hash write-only attribute", err.Error())
    return
}
resp.Diagnostics.Append(resp.Private.SetKey(ctx, apiTokenHasher.PrivateStateKey("api_token"), hash)...)
```

### After Delete (optional hygiene)

When the resource is destroyed, clear private-state keys for any write-only attributes that were tracked:

```go
resp.Diagnostics.Append(resp.Private.SetKey(ctx, apiTokenHasher.PrivateStateKey("api_token"), nil)...)
```

### Spot-check (helper API)

The example above uses only exported symbols from `internal/utils/writeonlyhash/hasher.go`. Reviewers can verify:

| Symbol | Signature |
| --- | --- |
| `Hasher` | struct type |
| `New` | `func New(resourceTypeName string) *Hasher` |
| `(*Hasher).Compute` | `func (h *Hasher) Compute(value string) ([]byte, error)` |
| `(*Hasher).Matches` | `func (h *Hasher) Matches(value string, storedHash []byte) bool` |
| `(*Hasher).PrivateStateKey` | `func (h *Hasher) PrivateStateKey(attributePath string) string` |

This worked example is **not** compiled as an `Example_*` test: a faithful `ModifyPlan` walkthrough requires Terraform Plugin Framework `resource.ModifyPlanRequest` types and would duplicate acceptance-test patterns. The table above plus unit tests in [`hasher_test.go`](../../internal/utils/writeonlyhash/hasher_test.go) are the compile-time guarantee for the helper itself.

## Anti-patterns

- **Logging the secret value** (including at debug/trace). Hash and compare only; never log `value` or `appliedToken`.
- **Including the value in diagnostics** (errors, warnings, or summaries). Name the attribute path only.
- **Using the helper from Read** to refresh or reconcile hashes. Read must not touch `secret_hash:*` private state.
- **Sharing one `Hasher` across resource types** or reusing another resource's salt string. Always `writeonlyhash.New("<this_resource_type>")` per resource.
- **Forgetting to clear the hash** when the write-only attribute or map element is removed from configuration.
- **Using SHA-256 or MD5** for offline-guessable secrets in state. Use this helper's bcrypt path instead.
- **Storing plaintext** in private state. Only store the `[]byte` returned by `Compute`.

## See also

- [`internal/utils/writeonlyhash`](../../internal/utils/writeonlyhash/) — helper implementation and [Godoc](https://pkg.go.dev/github.com/elastic/terraform-provider-elasticstack/internal/utils/writeonlyhash)
- [HashiCorp `random_password` — `bcrypt_hash`](https://registry.terraform.io/providers/hashicorp/random/latest/docs/resources/password#bcrypt_hash) — precedent for bcrypt hashing of secrets in Terraform state machinery
- [Content connector write-only requirements (REQ-011)](../../openspec/changes/elasticsearch-content-connector/specs/elasticsearch-content-connector/spec.md) — production adoption on `configuration_values["<key>"].secret_value` (moves to `openspec/specs/` after archive)
- [`coding-standards.md`](./coding-standards.md) — schema conventions, including the Write-only secret attributes subsection
