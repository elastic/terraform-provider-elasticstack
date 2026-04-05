variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title       = var.dashboard_title
  description = "Dashboard with Range Slider Control Panel (no config block)"

  time_range {
    from = "now-15m"
    to   = "now"
  }
  refresh_interval {
    pause = true
    value = 0
  }
  query {
    language = "kuery"
    text     = ""
  }
  panels = [{
    type = "range_slider_control"
    grid = {
      x = 0
      y = 0
      w = 24
      h = 4
    }
  }]
}
