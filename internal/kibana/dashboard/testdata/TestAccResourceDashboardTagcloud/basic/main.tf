variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title                  = var.dashboard_title
  description            = "Dashboard with Tagcloud Panel"
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
    tagcloud_config = {
      title       = "Sample Tagcloud"
      description = "Test tagcloud visualization"
      dataset_json = jsonencode({
        type = "dataView"
        id   = "metrics-*"
      })
      query = {
        language = "kuery"
        query    = ""
      }
      metric_json = jsonencode({
        operation = "count"
      })
      tag_by_json = jsonencode({
        operation = "terms"
        fields    = ["host.name"]
        size      = 10
      })
      orientation = "horizontal"
      font_size = {
        min = 18
        max = 72
      }
      ignore_global_filters = false
      sampling              = 1
    }
  }]
}
