variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title       = var.dashboard_title
  description = "Dashboard with Gauge Panel (ES|QL)"
  time_range = {
    from = "now-15m"
    to   = "now"
  }
  refresh_interval = {
    pause = true
    value = 0
  }
  query = {
    language = "kql"
    text     = ""
  }
  panels = [{
    type = "vis"
    grid = {
      x = 0
      y = 0
      w = 24
      h = 15
    }
    vis_config = {
      by_value = {
        gauge_config = {
          title       = "ESQL Gauge"
          description = "Gauge visualization using ES|QL"
          data_source_json = jsonencode({
            type  = "esql"
            query = "FROM metrics-* | STATS revenue = SUM(value) | LIMIT 1"
          })

          # Omit `query` for ES|QL mode.

          esql_metric = {
            column      = "revenue"
            format_json = jsonencode({ type = "number" })
            label       = "Revenue"
          }

          styling = {
            shape_json = jsonencode({
              type = "circle"
            })
          }

          ignore_global_filters = false
          sampling              = 1
        }
      }
    }
  }]
}
