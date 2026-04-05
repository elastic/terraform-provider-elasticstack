variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title       = var.dashboard_title
  description = "Dashboard with Time Slider Control Panel (empty config block)"

  time_from = "now-15m"
  time_to   = "now"

  refresh_interval_pause = true
  refresh_interval_value = 0

  query_language = "kql"
  query_text     = ""

  panels = [{
    type = "time_slider_control"
    grid = {
      x = 0
      y = 0
      w = 24
      h = 4
    }
    time_slider_control_config = {}
  }]
}
