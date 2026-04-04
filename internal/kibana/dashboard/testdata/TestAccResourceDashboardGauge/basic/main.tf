variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title                  = var.dashboard_title
  description            = "Dashboard with Gauge Panel"
  time_from              = "now-15m"
  time_to                = "now"
  refresh_interval_pause = true
  refresh_interval_value = 0
  query_language         = "kql"
  query_text             = ""

  panels = [{
    type = "lens"
    grid = {
      x = 0
      y = 0
      w = 24
      h = 15
    }
    gauge_config = {
      title       = "Sample Gauge"
      description = "Test gauge visualization"
      dataset_json = jsonencode({
        type  = "index"
        index = "metrics-*"

        time_field = "@timestamp"
      })
      query = {
        language   = "kql"
        expression = ""
      }
      metric_json = jsonencode({
        operation     = "count"
        empty_as_null = false
      })
      shape_json = jsonencode({
        type = "circle"
      })
      ignore_global_filters = false
      sampling              = 1
    }
  }]
}
