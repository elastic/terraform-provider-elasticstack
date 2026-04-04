# Bug: `elasticstack_elasticsearch_ml_datafeed_state` — `start` attribute remains unknown after create with `state = "stopped"`

## Description

When creating an `elasticstack_elasticsearch_ml_datafeed_state` resource with `state = "stopped"` and no explicit `start` time, Terraform reports:

> Provider returned invalid result object after apply. The provider still indicated an unknown value for `start`.

The resource is saved to state but is immediately **tainted**, causing every subsequent `terraform apply` to attempt recreation and hit the same error in a loop.

## Expected Behavior

Creating a datafeed state resource with `state = "stopped"` (and no explicit `start` time) should succeed cleanly. The `start` attribute should resolve to `null` since the datafeed is not running and has no start time.

## Actual Behavior

The `start` attribute remains unknown after apply. Terraform taints the resource. Subsequent applies attempt to replace the tainted resource, hitting the same error repeatedly.

**Workaround:** Run `terraform untaint <resource_address>` after the initial create.

## Steps to Reproduce

```hcl
resource "elasticstack_elasticsearch_ml_datafeed_state" "test" {
  datafeed_id = elasticstack_elasticsearch_ml_datafeed.test.datafeed_id
  state       = "stopped"
}
```

1. `terraform apply` — resource is created but tainted
2. `terraform state show <resource>` — `start` is unknown
3. Every subsequent `terraform apply` attempts to replace the tainted resource

## Root Cause Analysis

The `start` attribute is defined as `Optional + Computed` with a `timetypes.RFC3339` custom type and two plan modifiers:

```go
"start": schema.StringAttribute{
    CustomType: timetypes.RFC3339Type{},
    Optional:   true,
    Computed:   true,
    PlanModifiers: []planmodifier.String{
        stringplanmodifier.UseStateForUnknown(),
        SetUnknownIfStateHasChanges(),
    },
},
```

During **Create**, there is no prior state, so:
- `UseStateForUnknown()` has no state to copy — it returns early
- `SetUnknownIfStateHasChanges()` returns early because `req.State.Raw.IsNull()`

The plan value for `start` therefore remains **unknown** (computed, to be determined during apply).

During apply, the `update()` function calls `read()`, which calls `SetStartAndEndFromAPI()`. For a stopped datafeed, this method checks:

```go
if d.Start.IsUnknown() {
    d.Start = timetypes.NewRFC3339Null()
}
```

This *should* resolve the unknown to null. However, the resulting state still contains an unknown value for `start`, suggesting either:

1. The `timetypes.RFC3339` null value does not properly serialize through `state.Set()` for this attribute, or
2. The `state.Set()` operation fails silently (perhaps during conversion between the custom type and the framework's internal representation), leaving computed attributes in their initial unknown state, or
3. There is an interaction between the `timetypes.RFC3339Type{}` custom type and the Terraform Plugin Framework's state serialization that prevents proper null value propagation

## Environment

- Provider version: 0.13.1
- Terraform version: >= 1.0.0
- `terraform-plugin-framework`: v1.17.0
- `terraform-plugin-framework-timetypes`: v0.5.0

## Related

- See also: inconsistent `start` value when transitioning from stopped to started (separate issue)
