variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title       = var.dashboard_title
  description = "Dashboard with invalid time slider end percentage"

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
    type = "time_slider_control"
    grid = {
      x = 0
      y = 0
      w = 24
      h = 4
    }
    time_slider_control_config = {
      end_percentage_of_time_range = -0.1
    }
  }]
}
