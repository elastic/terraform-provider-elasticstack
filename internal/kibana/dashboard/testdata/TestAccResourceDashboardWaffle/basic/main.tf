variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title                  = var.dashboard_title
  description            = "Dashboard with Waffle Panel"
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
    type = "lens"
    grid = {
      x = 0
      y = 0
      w = 24
      h = 15
    }
    waffle_config = {
      title       = "Sample Waffle"
      description = "Test waffle visualization"
      dataset_json = jsonencode({
        type  = "index"
        index = "metrics-*"

        time_field = "@timestamp"
      })
      query = {
        language   = "kql"
        expression = ""
      }
      legend = {
        size   = "m"
        values = ["absolute"]
      }
      metrics = [
        {
          config = jsonencode({
            operation = "count"
          })
        }
      ]
      ignore_global_filters = false
      sampling              = 1
    }
  }]
}
