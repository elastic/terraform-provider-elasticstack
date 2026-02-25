variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title       = var.dashboard_title
  description = "Dashboard with Panels"

  time_range = {
    from = "now-15m"
    to   = "now"
  }

  refresh_interval = {
    pause = true
    value = 0
  }

  query = {
    language = "kuery"
    text     = ""
  }

  sections = [{
    title = "My Section"
    grid = {
      y = 0
    }
    panels = [{
      type = "DASHBOARD_MARKDOWN"
      grid = {
        x = 0
        y = 0
        w = 24
        h = 10
      }
      markdown_config = {
        content    = "First markdown panel"
        title      = "My First Markdown Panel"
        hide_title = false
      }
    }]
  }]
}
