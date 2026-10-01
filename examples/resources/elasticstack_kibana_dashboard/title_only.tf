// Example: minimal dashboard. Only `title` is required; `time_range`, `refresh_interval`, and `query` are optional.
// Kibana applies no defaults for omitted blocks. Removing a block later clears it rather than resetting it to a default.

resource "elasticstack_kibana_dashboard" "title_only" {
  title = "Title-only dashboard"
}
