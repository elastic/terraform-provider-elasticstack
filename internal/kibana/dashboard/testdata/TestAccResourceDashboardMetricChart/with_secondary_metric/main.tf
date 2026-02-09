variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title                  = var.dashboard_title
  description            = "Dashboard with Metric Chart Panel with Secondary Metric"
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
      title       = "Sample Metric Chart with Secondary Metric"
      description = "Test metric chart with secondary metric"
      dataset = jsonencode({
        type = "dataView"
        id   = "metrics-*"
      })
      query = {
        language = "kuery"
        query    = "status:active"
      }
      metrics = [
        {
          config = jsonencode({
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
        },
        {
          config = jsonencode({
            type  = "secondary",
            operation = "last_value",
            field = "@timestamp",
            sort_by = "@timestamp",
            show_array_values = false,
            filter = {
              query = "\"@timestamp\": *"
              language = "kuery"
            }
          })
        }
      ]
      ignore_global_filters = false
      sampling              = 1
    }
  }]
}
