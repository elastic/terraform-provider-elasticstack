Advanced settings to manage, keyed by setting name (for example `dateFormat:tz`). Each value must be a JSON-encoded string so that the setting type is preserved, for example `jsonencode("Browser")`, `jsonencode(true)`, `jsonencode(100)` or `jsonencode(["a", "b"])`. Removing a setting from this map resets it to its Kibana default.

Settings whose Kibana type is `json` (for example `timepicker:timeDefaults`) are stored by Kibana as JSON strings, so encode them twice: `jsonencode(jsonencode({ from = "now-30m", to = "now" }))`.
