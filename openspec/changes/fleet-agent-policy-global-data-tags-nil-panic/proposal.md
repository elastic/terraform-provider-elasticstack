## Why

The `elasticstack_fleet_agent_policy` resource panics (SIGSEGV) during `Update` when a
`global_data_tags` map entry has **neither** `string_value` nor `number_value` set. The
crash occurs in `internal/fleet/agentpolicy/models.go` inside `convertGlobalDataTags`
because `item.StringValue.ValueStringPointer()` returns `nil` when `StringValue` is null,
and the `else` branch unconditionally dereferences `item.NumberValue.ValueFloat32Pointer()`
which is also nil when `NumberValue` is null.

The existing schema only attaches `ConflictsWith` validators (preventing *both* values from
being set simultaneously) but has no constraint preventing *neither* from being set. A user
can write:

```hcl
global_data_tags = {
  "my_tag" = {}
}
```

Terraform accepts this at plan time and the provider crashes at apply time. Reports confirm
this affects all provider versions since `global_data_tags` was introduced (v0.12.x cycle).
The bug is not Elastic Stack version-specific; ES 9.2.x is merely when a user first
triggered it via an `apply`.

## What Changes

- **Runtime guard** in `convertGlobalDataTags` (`internal/fleet/agentpolicy/models.go`):
  Replace the unconditional nil-dereference pointer path with an explicit null/unknown check
  that returns a descriptive diagnostic error instead of panicking when both fields are null.

- **Schema-level validator** (`internal/fleet/agentpolicy/schema.go`): Add
  `stringvalidator.AtLeastOneOf` / `float32validator.AtLeastOneOf` to both `string_value`
  and `number_value` so that an empty `{}` entry is rejected at `terraform plan` time with a
  clear error message. Both validators are available in
  `github.com/hashicorp/terraform-plugin-framework-validators` v0.19.0, which is already a
  declared dependency.

- **Unit test** in `internal/fleet/agentpolicy/models_test.go`: Assert that
  `convertGlobalDataTags` returns a diagnostic error (not a panic) when an item has both
  fields null.

## Capabilities

### New Capabilities

- _(none)_

### Modified Capabilities

- **`elasticstack_fleet_agent_policy`** (`global_data_tags`): Null-null tag entries now
  produce a diagnostic error rather than a provider panic; additionally rejected at plan time
  by new schema validators.

## Impact

- **Users**: The panic is eliminated. Users with `global_data_tags = { "x" = {} }` in their
  configuration (which was always semantically invalid) will now see a clear plan-time
  validation error instead of a runtime crash on apply. Users with valid configurations are
  unaffected.
- **Code**: Changes are isolated to `internal/fleet/agentpolicy/models.go`,
  `internal/fleet/agentpolicy/schema.go`, and `internal/fleet/agentpolicy/models_test.go`.
- **No state migration required**: The fix does not change the state schema version or stored
  state shape. Existing valid state continues to work without modification.
