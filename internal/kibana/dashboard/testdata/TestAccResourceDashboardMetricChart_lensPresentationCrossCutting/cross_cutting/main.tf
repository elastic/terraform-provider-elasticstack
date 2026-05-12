variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title       = var.dashboard_title
  description = "Dashboard with metric chart presentation acceptance"
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
    metric_chart_config = {
      title       = "Metric presentation acc"
      description = "Cross-cutting lens presentation acceptance"
      time_range = {
        from = "now-30d"
        to   = "now-1d"
      }
      hide_title = true
      drilldowns = [
        {
          discover_drilldown = {
            label = "Open Discover"
          }
        }
      ]
      data_source_json = jsonencode({
        type          = "data_view_spec"
        index_pattern = "metrics-*"
        time_field    = "@timestamp"
      })
      query = {
        language   = "kql"
        expression = ""
      }
      metrics = [
        {
          config_json = jsonencode({
            type      = "primary"
            operation = "count"
            format = {
              type = "number"
            }
          })
        }
      ]
      ignore_global_filters = false
      sampling              = 1
    }
  }]
}
