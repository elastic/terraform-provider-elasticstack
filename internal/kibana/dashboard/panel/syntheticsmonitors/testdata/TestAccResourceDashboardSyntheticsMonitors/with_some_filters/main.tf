variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title       = var.dashboard_title
  description = "Dashboard with Synthetics Monitors Panel with some filters"

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
    type = "synthetics_monitors"
    grid = {
      x = 0
      y = 0
      w = 24
      h = 15
    }
    synthetics_monitors_config = {
      title       = "Synthetics Monitors"
      description = "Shows the production monitors"
      hide_title  = true
      hide_border = false
      view        = "compactView"
      filters = {
        projects = [
          { label = "My Project", value = "my-project" }
        ]
        tags = [
          { label = "production", value = "production" }
        ]
      }
    }
  }]
}
