variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title                  = var.dashboard_title
  description            = "Dashboard with Legacy Metric Panel"
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
      h = 10
    }
    legacy_metric_config = {
      title       = "Legacy Metric"
      description = "Legacy metric chart"
      dataset_json = jsonencode({
        type  = "index"
        index = "metrics-*"

        time_field = "@timestamp"
      })
      query = {
        language   = "kql"
        expression = ""
      }
      filters = [
        {
          filter_json = jsonencode({
            type = "condition"
            condition = {
              field    = "status"
              operator = "is"
              value    = "200"
            }
          })
        }
      ]
      metric_json = jsonencode({
        operation     = "count"
        empty_as_null = false
        format = {
          type     = "number"
          decimals = 2
          compact  = false
        }
      })
      sampling              = 0.5
      ignore_global_filters = true
    }
  }]
}
