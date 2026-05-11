resource "elasticstack_kibana_dashboard" "my_dashboard" {
  title       = "My Dashboard"
  description = "A dashboard showing key metrics"

  time_range = {
    from = "now-15m"
    to   = "now"
  }

  refresh_interval = {
    pause = false
    value = 60000 # 60 seconds
  }

  query = {
    language = "kql"
    text     = "status:success"
  }

  # Optional tags
  tags = ["production", "monitoring"]
}

# Example with JSON query (mutually exclusive with query.text)
resource "elasticstack_kibana_dashboard" "my_dashboard_json" {
  title       = "My Dashboard with JSON Query"
  description = "A dashboard with a structured query"

  time_range = {
    from = "now-15m"
    to   = "now"
  }

  refresh_interval = {
    pause = false
    value = 60000 # 60 seconds
  }

  query = {
    language = "kql"
    json = jsonencode({
      bool = {
        must = [
          {
            match = {
              status = "success"
            }
          }
        ]
      }
    })
  }

  # Optional tags
  tags = ["production", "monitoring"]
}

# Inline Lens (`lens-dashboard-app`) panels: typed by-value metric charts (`metric_chart_config`, etc.;
# not `by_value.config_json`). These blocks intentionally omit REQ-037 presentation fields; use `type = "vis"`
# panels with typed chart `*_config` blocks when you need `time_range`, `drilldowns`, etc.
resource "elasticstack_kibana_dashboard" "lens_app_typed_by_value" {
  title            = "Dashboard with lens-dashboard-app (typed by-value)"
  description      = "Example: two metric panels sharing dashboard time_range (no chart-level presentation fields)"
  time_range       = { from = "now-15m", to = "now" }
  refresh_interval = { pause = true, value = 0 }
  query            = { language = "kql", text = "" }

  panels = [
    {
      type = "lens-dashboard-app"
      grid = { x = 0, y = 0, w = 12, h = 15 }
      lens_dashboard_app_config = {
        by_value = {
          metric_chart_config = {
            data_source_json = jsonencode({
              type          = "data_view_spec"
              index_pattern = "metrics-*"
              time_field    = "@timestamp"
            })
            query = { expression = "" }
            metrics = [{
              config_json = jsonencode({
                type      = "primary"
                operation = "count"
                format    = { type = "number" }
              })
            }]
          }
        }
      }
    },
    {
      type = "lens-dashboard-app"
      grid = { x = 12, y = 0, w = 12, h = 15 }
      lens_dashboard_app_config = {
        by_value = {
          metric_chart_config = {
            data_source_json = jsonencode({
              type          = "data_view_spec"
              index_pattern = "metrics-*"
              time_field    = "@timestamp"
            })
            query = { expression = "" }
            metrics = [{
              config_json = jsonencode({
                type      = "primary"
                operation = "count"
                format    = { type = "number" }
              })
            }]
          }
        }
      }
    }
  ]
}
