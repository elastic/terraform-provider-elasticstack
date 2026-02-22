variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title                  = var.dashboard_title
  description            = "Dashboard with Legacy Metric Panel"
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
      h = 10
    }
    legacy_metric_config = {
      title       = "Legacy Metric"
      description = "Legacy metric chart"
      dataset_json = jsonencode({
        type = "dataView"
        id   = "metrics-*"
      })
      query = {
        language = "kuery"
        query    = ""
      }
      filters = [
        {
          query = "status:200"
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
