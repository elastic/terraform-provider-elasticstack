variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title                  = var.dashboard_title
  description            = "Dashboard with Treemap Panel (ES|QL)"
  time_from              = "now-15m"
  time_to                = "now"
  refresh_interval_pause = true
  refresh_interval_value = 0
  query_language         = "kuery"
  query_text             = ""

  panels = [{
    type = "lens"
    grid = {
      x = 0
      y = 0
      w = 24
      h = 15
    }

    treemap_config = {
      title       = "ESQL Treemap"
      description = "Treemap visualization using ES|QL"

      dataset_json = jsonencode({
        type  = "esql"
        query = "FROM metrics-* | KEEP host.name, bytes | LIMIT 50"
      })

      # Note: omit `query` block for ES|QL mode.

      group_by_json = jsonencode([
        {
          operation   = "value"
          column      = "host.name"
          collapse_by = "avg"
        }
      ])

      metrics_json = jsonencode([
        {
          operation = "value"
          column    = "bytes"
          color = {
            type  = "static"
            color = "#54B399"
          }
        }
      ])

      label_position = "hidden"

      legend = {
        nested               = false
        size                 = "small"
        visible              = "show"
        truncate_after_lines = 10
      }

      value_display = {
        mode = "absolute"
      }

      ignore_global_filters = true
      sampling              = 0.5
    }
  }]
}
