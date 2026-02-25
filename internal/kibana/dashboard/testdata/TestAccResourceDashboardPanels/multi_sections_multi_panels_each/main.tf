variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title       = var.dashboard_title
  description = "Dashboard with Multiple Sections (Multiple Panels Each)"

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
        title      = "Section One Panel One"
        hide_title = false
      }
      }, {
      type = "DASHBOARD_MARKDOWN"
      grid = {
        x = 0
        y = 10
        w = 24
        h = 10
      }
      markdown_config = {
        content    = "Section one - panel two"
        title      = "Section One Panel Two"
        hide_title = false
      }
    }]
    }, {
    title = "Section Two"
    grid = {
      y = 20
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
        title      = "Section Two Panel One"
        hide_title = false
      }
      }, {
      type = "DASHBOARD_MARKDOWN"
      grid = {
        x = 0
        y = 10
        w = 24
        h = 10
      }
      markdown_config = {
        content    = "Section two - panel two"
        title      = "Section Two Panel Two"
        hide_title = false
      }
    }]
  }]
}
