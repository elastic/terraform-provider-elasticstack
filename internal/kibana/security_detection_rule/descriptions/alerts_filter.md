Conditions that affect whether the action runs. If you specify multiple conditions, all conditions must be met for the action to run.

The `query` attribute accepts a KQL string and a JSON array of Kibana filter objects. Use `filters_json = jsonencode([])` when no filters are required. Example: `alerts_filter = { query = { kql = "event.action : \"test\"" filters_json = jsonencode([]) } }`.
