variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title                  = var.dashboard_title
  description            = "Dashboard with Metric Chart Panel with Filters"
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
    metric_chart_config = {
      title       = "Sample Metric Chart with Filters"
      description = "Test metric chart with filters visualization"
      dataset_json = jsonencode({
        type = "dataView"
        id   = "metrics-*"
      })
      query = {
        language = "kuery"
        query    = "status:active"
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
      breakdown_by_json = jsonencode({
        operation = "terms"
        fields    = ["category"]
        limit     = 3
        rank_by = {
          direction = "desc"
          metric    = 0
          type      = "column"
        }
      })
      filters = [
        {
          filter_json = jsonencode({
            type = "condition"
            condition = {
              field    = "event.category"
              operator = "is"
              value    = "web"
            }
          })
        }
      ]
      ignore_global_filters = false
      sampling              = 1
    }
  }]
}
