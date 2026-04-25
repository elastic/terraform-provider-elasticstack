variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title       = var.dashboard_title
  description = "lens-dashboard-app by_value acceptance (REQ-035 / enrich preservation)"
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
    type = "lens-dashboard-app"
    grid = {
      x = 0
      y = 0
      w = 24
      h = 10
    }
    lens_dashboard_app_config = {
      by_value = {
        config_json = jsonencode({
          type  = "metric"
          title = "Acc by-value"
          data_source = {
            type          = "data_view_spec"
            index_pattern = "metrics-*"
            time_field    = "@timestamp"
          }
          filters = []
          # API requires at least one metric; align with `metric_chart_config` shape.
          metrics = [
            {
              type      = "primary"
              operation = "count"
              format    = { type = "number" }
            }
          ]
          query = {
            language   = "kql"
            expression = ""
          }
          styling = {
            icon = { name = "heart" }
          }
          time_range = {
            from = "now-15m"
            to   = "now"
          }
        })
      }
    }
  }]
}
