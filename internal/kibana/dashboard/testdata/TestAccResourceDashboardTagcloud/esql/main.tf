variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title       = var.dashboard_title
  description = "Dashboard with Tagcloud Panel (ES|QL)"
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
        tagcloud_config = {
          title       = "ESQL Tagcloud"
          description = "Tagcloud visualization using ES|QL"
          data_source_json = jsonencode({
            type  = "esql"
            query = "FROM logs-* | STATS count = COUNT() BY host | LIMIT 50"
          })

          # Omit `query` for ES|QL mode.

          esql_metric = {
            column      = "count"
            format_json = jsonencode({ type = "number" })
          }

          esql_tag_by = {
            column      = "host"
            format_json = jsonencode({ type = "number" })
            color_json = jsonencode({
              mode    = "categorical"
              palette = "default"
              mapping = []
            })
          }

          orientation = "horizontal"
          font_size = {
            min = 18
            max = 72
          }

          ignore_global_filters = false
          sampling              = 1
        }
      }
    }
  }]
}
