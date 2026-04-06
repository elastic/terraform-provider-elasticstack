variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title       = var.dashboard_title
  description = "Dashboard with Waffle Panel"
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
      title       = "Complete Waffle"
      description = "Complete waffle visualization"
      dataset_json = jsonencode({
        type  = "index"
        index = "metrics-*"

        time_field = "@timestamp"
      })
      query = {
        language   = "kql"
        expression = ""
      }
      filters = [
        {
          filter_json = jsonencode({
            type = "condition"
            condition = {
              field    = "host.os.keyword"
              operator = "is"
              value    = "linux"
            }
          })
        }
      ]
      legend = {
        size                 = "s"
        visible              = "visible"
        truncate_after_lines = 8
        values               = ["absolute"]
      }
      value_display = {
        mode             = "percentage"
        percent_decimals = 1
      }
      metrics = [
        {
          config = jsonencode({
            operation = "count"
          })
        }
      ]
      ignore_global_filters = true
      sampling              = 0.5
    }
  }]
}
