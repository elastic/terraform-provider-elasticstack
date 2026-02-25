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
    language = "kuery"
    text     = ""
  }

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
            alignments = {
              labels = "center"
            }
            icon = {
              name = "document"
            }
          })
        }
      ]
      breakdown_by_json = jsonencode({
        operation = "terms"
        fields    = ["category"]
        size      = 3
        columns   = 3
        rank_by = {
          direction = "desc"
          metric    = 0
          type      = "column"
        }
      })
      filters = [
        {
          language = "kuery"
          query    = "event.category:web"
        }
      ]
      ignore_global_filters = false
      sampling              = 1
    }
  }]
}
