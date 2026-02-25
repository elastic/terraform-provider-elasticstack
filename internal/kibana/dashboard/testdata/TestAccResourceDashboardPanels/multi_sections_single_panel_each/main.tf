variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title       = var.dashboard_title
  description = "Dashboard with Multiple Sections (Single Panel Each)"

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
    title = "Section One"
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
        content    = "Section one - panel one"
        title      = "Section One Panel"
        hide_title = false
      }
    }]
    }, {
    title = "Section Two"
    grid = {
      y = 10
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
        content    = "Section two - panel one"
        title      = "Section Two Panel"
        hide_title = false
      }
    }]
  }]
}
