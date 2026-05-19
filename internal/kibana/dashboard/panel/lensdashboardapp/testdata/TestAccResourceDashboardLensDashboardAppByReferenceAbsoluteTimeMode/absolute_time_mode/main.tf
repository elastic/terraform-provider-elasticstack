variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title       = var.dashboard_title
  description = "lens-dashboard-app by_reference absolute mode"
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
    type = "lens-dashboard-app"
    grid = {
      x = 0
      y = 0
      w = 24
      h = 10
    }
    lens_dashboard_app_config = {
      by_reference = {
        ref_id = "lensRef2"
        time_range = {
          from = "2024-06-01T00:00:00.000Z"
          to   = "2024-06-01T12:00:00.000Z"
          mode = "absolute"
        }
      }
    }
  }]
}
