Conditions that affect whether the action runs. If you specify multiple conditions, all conditions must be met for the action to run.

The `query` block supports a KQL string and a JSON array of Kibana filter objects. Use `filters_json = jsonencode([])` when no filters are required.

Example:

```hcl
alerts_filter {
  query {
    kql          = "event.action : \"test\""
    filters_json = jsonencode([])
  }
  timeframe {
    days        = [1, 2, 3, 4, 5]
    timezone    = "UTC"
    hours_start = "08:00"
    hours_end   = "17:00"
  }
}
```
