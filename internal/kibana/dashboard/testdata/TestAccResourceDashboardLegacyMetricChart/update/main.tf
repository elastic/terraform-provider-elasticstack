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
      title       = "Updated Legacy Metric"
      description = "Updated description"
      dataset = jsonencode({
        type = "dataView"
        id   = "metrics-*"
      })
      query = {
        language = "lucene"
        query    = "status:500"
      }
      filters = [
        {
          query = "status:200"
        }
      ]
      metric = jsonencode({
        operation     = "count"
        empty_as_null = false
        format = {
          type     = "number"
          decimals = 2
          compact  = false
        }
      })
      sampling              = 1
      ignore_global_filters = false
    }
  }]
}
