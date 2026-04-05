variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title                  = var.dashboard_title
  description            = "Dashboard with Tagcloud Panel and Filters"
  time_from              = "now-1h"
  time_to                = "now"
  refresh_interval_pause = false
  refresh_interval_value = 30000
  query_language         = "kql"
  query_text             = ""

  panels = [{
    type = "lens"
    grid = {
      x = 0
      y = 0
      w = 48
      h = 20
    }
    tagcloud_config = {
      title       = "Filtered Tagcloud"
      description = "Tagcloud with filters and custom settings"
      dataset_json = jsonencode({
        type  = "index"
        index = "logs-*"

        time_field = "@timestamp"
      })
      query = {
        language   = "kql"
        expression = "service.name:*"
      }
      filters = [
        {
          filter_json = jsonencode({
            type = "condition"
            condition = {
              field    = "log.level"
              operator = "is"
              value    = "error"
            }
          })
        }
      ]
      metric_json = jsonencode({
        operation = "sum"
        field     = "event.duration"
      })
      tag_by_json = jsonencode({
        operation = "terms"
        fields    = ["service.name"]
        limit     = 15
      })
      orientation           = "vertical"
      ignore_global_filters = true
      sampling              = 0.5
      font_size = {
        min = 12
        max = 100
      }
    }
  }]
}
