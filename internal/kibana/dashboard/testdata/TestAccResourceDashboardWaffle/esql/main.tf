variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title                  = var.dashboard_title
  description            = "Dashboard with Waffle Panel (ES|QL)"
  time_range {
    from = "now-15m"
    to   = "now"
  }
  refresh_interval {
    pause = true
    value = 0
  }
  query {
    language = "kql"
    text     = ""
  }
  panels = [{
    type = "lens"
    grid = {
      x = 0
      y = 0
      w = 24
      h = 15
    }

    waffle_config = {
      title       = "ESQL Waffle"
      description = "Waffle visualization using ES|QL"

      dataset_json = jsonencode({
        type = "esql"
        # Single-bucket STATS avoids group-by coloring rules in Lens ("Coloring cannot be
        # assigned to a collapsed grouping dimension" with BY + collapse).
        query = "FROM metrics-* | STATS c = COUNT() | LIMIT 10"
      })

      # Omit `query` for ES|QL mode (see provider docs).

      legend = {
        size = "m"
      }

      esql_metrics = [{
        column      = "c"
        format_json = jsonencode({ type = "number" })
        color = {
          type  = "static"
          color = "#006BB4"
        }
      }]

      ignore_global_filters = false
      sampling              = 1
    }
  }]
}
