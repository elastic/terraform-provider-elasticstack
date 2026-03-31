variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title                  = var.dashboard_title
  description            = "Dashboard with Mosaic Panel (ES|QL)"
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

    mosaic_config = {
      title       = "ESQL Mosaic"
      description = "Mosaic visualization using ES|QL"

      dataset_json = jsonencode({
        type  = "esql"
        query = "FROM metrics-* | KEEP host.name, service.name, bytes | LIMIT 50"
      })

      # Note: omit `query` block for ES|QL mode.

      group_by_json = jsonencode([
        {
          operation   = "value"
          column      = "host.name"
          collapse_by = "avg"
          color = {
            mode    = "categorical"
            palette = "default"
            mapping = []
            unassigned = {
              type  = "color_code"
              value = "#D3DAE6"
            }
          }
        }
      ])

      group_breakdown_by_json = jsonencode([
        {
          operation   = "value"
          column      = "service.name"
          collapse_by = "avg"
          color = {
            mode    = "categorical"
            palette = "default"
            mapping = []
            unassigned = {
              type  = "color_code"
              value = "#D3DAE6"
            }
          }
        }
      ])

      metrics_json = jsonencode([
        {
          operation = "value"
          column    = "bytes"
          format = {
            type     = "number"
            decimals = 2
          }
        }
      ])

      legend = {
        nested               = false
        size                 = "small"
        visible              = "visible"
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
