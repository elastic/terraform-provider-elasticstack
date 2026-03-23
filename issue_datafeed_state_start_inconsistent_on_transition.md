# Bug: `elasticstack_elasticsearch_ml_datafeed_state` — inconsistent `start` value when transitioning from stopped to started

## Description

When updating an `elasticstack_elasticsearch_ml_datafeed_state` resource from `state = "stopped"` to `state = "started"` (without an explicit `start` time), Terraform reports:

> Provider produced inconsistent result after apply. `.start`: was null, but now `cty.StringVal("2026-03-10T02:40:04+13:00")`

The datafeed **does** start successfully — the error is purely a plan/result mismatch. The resource may or may not be tainted depending on the Terraform version.

## Expected Behavior

The plan should predict that `start` will be `(known after apply)` (unknown) when the `state` attribute changes, since starting a datafeed in real-time mode produces a server-determined start time. After apply, the concrete start timestamp should be accepted by Terraform as a valid resolution of the unknown plan value.

## Actual Behavior

The plan predicts `start = null` (carried forward from the stopped state). When the provider returns the actual start timestamp from the Elasticsearch API, Terraform rejects it as an inconsistent result.

## Steps to Reproduce

1. Create the datafeed state with `state = "stopped"`:
```hcl
resource "elasticstack_elasticsearch_ml_datafeed_state" "test" {
  datafeed_id = elasticstack_elasticsearch_ml_datafeed.test.datafeed_id
  state       = "stopped"
}
```

2. Update to `state = "started"`:
```hcl
resource "elasticstack_elasticsearch_ml_datafeed_state" "test" {
  datafeed_id = elasticstack_elasticsearch_ml_datafeed.test.datafeed_id
  state       = "started"
}
```

3. `terraform apply` — error: `.start` was null, but now has a concrete value

## Root Cause Analysis

The `start` attribute uses a custom plan modifier `SetUnknownIfStateHasChanges()` that is designed to mark `start` as unknown whenever the `state` attribute changes:

```go
func (s setUnknownIfStateHasChanges) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
    if req.State.Raw.IsNull() || req.Config.Raw.IsNull() {
        return
    }
    if utils.IsKnown(req.ConfigValue) {
        return
    }

    var stateValue, configValue types.String
    resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("state"), &stateValue)...)
    resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("state"), &configValue)...)

    if !stateValue.Equal(configValue) {
        resp.PlanValue = types.StringUnknown()  // <-- Sets base type unknown
    }
}
```

The modifier correctly detects the state change (stopped → started) and sets `resp.PlanValue = types.StringUnknown()`. However, the `start` attribute uses `CustomType: timetypes.RFC3339Type{}`.

The issue is likely that the plan modifier sets a **base type** unknown (`types.StringUnknown()`) rather than a **custom type** unknown. The Terraform Plugin Framework should convert this via `RFC3339Type{}.ValueFromString()`, but there may be a subtle interaction where:

1. The framework does not properly convert the base-type unknown value back to the custom type after plan modification, causing the plan to retain the previous value (null from `UseStateForUnknown()`), or
2. The conversion happens but the resulting plan value is compared against the prior state value using custom type semantics that treat the unknown differently

The plan modifier chain is:
1. `UseStateForUnknown()` — copies state value (null, from stopped state) into plan
2. `SetUnknownIfStateHasChanges()` — should override with unknown since state changed

If step 2's output is silently discarded due to a type mismatch, the plan retains `start = null` from step 1. When the provider returns the actual start timestamp, Terraform rejects it.

### Secondary issue in `updateAfterMissedTransition()`

There is also a code path in `updateAfterMissedTransition()` that unconditionally sets `start` to null if unknown, without reading the actual start time from the API:

```go
if data.Start.IsUnknown() {
    data.Start = timetypes.NewRFC3339Null()
}
return &data, nil
```

This runs when a started datafeed stops too quickly for the wait function to detect. The method should read start/end values from the datafeed stats instead of defaulting to null, since the datafeed may have actually started and acquired a real start time before stopping.

## Environment

- Provider version: 0.13.1
- Terraform version: >= 1.0.0
- `terraform-plugin-framework`: v1.17.0
- `terraform-plugin-framework-timetypes`: v0.5.0

## Related

- See also: `start` remains unknown after create with stopped state (separate issue)
