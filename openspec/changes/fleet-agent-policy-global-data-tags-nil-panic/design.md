## Context

The `global_data_tags` attribute of `elasticstack_fleet_agent_policy` maps tag names to a
nested object with exactly one of `string_value` (string) or `number_value` (float32) set.
The conversion from Terraform model to Kibana API model happens in
`(*agentPolicyModel).convertGlobalDataTags` in
[`internal/fleet/agentpolicy/models.go`](../../../internal/fleet/agentpolicy/models.go).

The current conversion logic (line ~291):

```go
if item.StringValue.ValueStringPointer() != nil {
    err = value.FromAgentPolicyGlobalDataTagsItemValue0(*item.StringValue.ValueStringPointer())
} else {
    // BUG: item.NumberValue may also be null → nil dereference
    err = value.FromAgentPolicyGlobalDataTagsItemValue1(*item.NumberValue.ValueFloat32Pointer())
}
```

`ValueStringPointer()` returns `nil` for a null `types.String`. The `else` branch then
dereferences `ValueFloat32Pointer()` which is also nil for a null `types.Float32`, causing
the SIGSEGV.

**Root trigger**: The `global_data_tags` schema
([`internal/fleet/agentpolicy/schema.go`](../../../internal/fleet/agentpolicy/schema.go),
lines ~164–178) marks both `string_value` and `number_value` as `Optional` only, with no
`AtLeastOneOf` constraint. The `ConflictsWith` validator only blocks *both* being set; it
does not block *neither* being set.

## Goals / Non-Goals

**Goals:**

- Eliminate the nil pointer dereference panic in `convertGlobalDataTags`.
- Reject invalid null-null `global_data_tags` entries at **plan time** with a clear user-
  facing error.
- Add a unit test proving `convertGlobalDataTags` emits diagnostics (not panics) for
  null-null items.

**Non-goals:**

- Adding a third value type (e.g., boolean) — not required by the current API spec.
- State migration — the state schema shape is unchanged.
- Changes to other Fleet resources.
- Backport guidance (that belongs to release management).

## Decisions

### Approach A — Defensive nil check (runtime guard)

Replace the pointer-dereference pattern with an explicit null/unknown guard:

```go
if !item.StringValue.IsNull() && !item.StringValue.IsUnknown() {
    err = value.FromAgentPolicyGlobalDataTagsItemValue0(item.StringValue.ValueString())
} else if !item.NumberValue.IsNull() && !item.NumberValue.IsUnknown() {
    err = value.FromAgentPolicyGlobalDataTagsItemValue1(item.NumberValue.ValueFloat32())
} else {
    diags.AddAttributeError(
        meta.Path,
        "Invalid global_data_tags entry",
        "Each entry in global_data_tags must have exactly one of string_value or number_value set.",
    )
    return kbapi.AgentPolicyGlobalDataTagsItem{}
}
```

This is the **safety net**: eliminates the crash for any path by which a null-null item
reaches the conversion function (config error, state corruption, API contract change).

### Approach B — Schema `AtLeastOneOf` validator (plan-time guard)

Add `stringvalidator.AtLeastOneOf` and `float32validator.AtLeastOneOf` on both attributes
inside the `global_data_tags` nested object:

```go
"string_value": schema.StringAttribute{
    Optional: true,
    Validators: []validator.String{
        stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("number_value")),
        stringvalidator.AtLeastOneOf(
            path.MatchRelative().AtParent().AtName("string_value"),
            path.MatchRelative().AtParent().AtName("number_value"),
        ),
    },
},
"number_value": schema.Float32Attribute{
    Optional: true,
    Validators: []validator.Float32{
        float32validator.ConflictsWith(path.MatchRelative().AtParent().AtName("string_value")),
        float32validator.AtLeastOneOf(
            path.MatchRelative().AtParent().AtName("string_value"),
            path.MatchRelative().AtParent().AtName("number_value"),
        ),
    },
},
```

**Decision: implement both A and B together.** Defense in depth is correct here:
- The schema validator (B) catches the common case (user misconfiguration) at plan time.
- The runtime guard (A) catches the edge case where null-null items arrive via state
  deserialization or a hypothetical API inconsistency, and converts a crash into a
  diagnostic error.

Using `ValueString()` / `ValueFloat32()` (not pointer variants) in Approach A avoids any
future pointer-nil risk after the `IsNull()` / `IsUnknown()` guard.

## Risks / Trade-offs

- **Schema tightening**: If a user has valid-looking state that somehow contains null-null
  entries (e.g., from a prior state import), `terraform plan` will now produce a validation
  error instead of a crash on apply. This is strictly better UX — a clear error vs. a panic.
- **Validator duplication**: `AtLeastOneOf` appears on both `string_value` and `number_value`
  (each referencing both paths). This is intentional and matches the existing `ConflictsWith`
  pattern used in this codebase. The framework deduplicates the actual validation checks at
  evaluation time.

## Open Questions

- Has the user confirmed their HCL config has a `global_data_tags` entry with neither
  `string_value` nor `number_value`? If so, the above analysis is complete. If not, it would
  be worth asking them to share their configuration.
- Could the Kibana Fleet API in 9.2.x return `global_data_tags` values of a type not
  parseable as float32 or string (e.g., boolean)? The current `populateFromAPI` error path
  would set `StringValue` to an empty-string fallback in that case (not null), so it wouldn't
  trigger the panic — but it would silently corrupt the data. If Kibana 9.2.x added boolean
  tag support, a separate schema extension would be needed.
- Is a patch backport to v0.12.x needed, or is a fix on main sufficient?
