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
  query_language         = "kuery"
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
      dataset = jsonencode({
        type = "dataView"
        id   = "logs-*"
      })
      query = {
        language = "kuery"
        query    = "service.name:*"
      }
      filters = [
        {
          query = "log.level:error OR log.level:warning"
        }
      ]
      metric = jsonencode({
        operation = "sum"
        field     = "event.duration"
      })
      tag_by = jsonencode({
        operation = "terms"
        fields    = ["service.name"]
        size      = 15
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
