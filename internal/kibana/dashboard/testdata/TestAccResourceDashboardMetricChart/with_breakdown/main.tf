variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title       = var.dashboard_title
  description = "Dashboard with Metric Chart Panel with Filters"
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
      title       = "Sample Metric Chart with Filters"
      description = "Test metric chart with filters visualization"
      data_source_json = jsonencode({
        type          = "data_view_spec"
        index_pattern = "metrics-*"

        time_field = "@timestamp"
      })
      query = {
        language   = "kql"
        expression = "status:active"
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
          direction    = "desc"
          metric_index = 0
          type         = "metric"
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
