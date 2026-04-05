variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title       = var.dashboard_title
  description = "Dashboard with time slider plus unsupported config_json"

  time_range {
    from = "now-15m"
    to   = "now"
  }
  refresh_interval {
    pause = true
    value = 0
  }
  query {
    language = "kql"
    text     = ""
  }
  panels = [{
    type = "time_slider_control"
    grid = {
      x = 0
      y = 0
      w = 24
      h = 4
    }
    config_json = jsonencode({
      start_percentage_of_time_range = 0.25
      end_percentage_of_time_range   = 0.75
    })
  }]
}
