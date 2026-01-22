variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title                  = var.dashboard_title
  description            = "Dashboard with Multiple Sections (Multiple Panels Each)"
  time_from              = "now-15m"
  time_to                = "now"
  refresh_interval_pause = true
  refresh_interval_value = 0
  query_language         = "kuery"
  query_text             = ""

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
      embeddable_config = {
        content           = "Section one - panel one"
        title             = "Section One Panel One"
        hide_panel_titles = false
      }
      }, {
      type = "DASHBOARD_MARKDOWN"
      grid = {
        x = 0
        y = 10
        w = 24
        h = 10
      }
      embeddable_config = {
        content           = "Section one - panel two"
        title             = "Section One Panel Two"
        hide_panel_titles = false
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
      embeddable_config = {
        content           = "Section two - panel one"
        title             = "Section Two Panel One"
        hide_panel_titles = false
      }
      }, {
      type = "DASHBOARD_MARKDOWN"
      grid = {
        x = 0
        y = 10
        w = 24
        h = 10
      }
      embeddable_config = {
        content           = "Section two - panel two"
        title             = "Section Two Panel Two"
        hide_panel_titles = false
      }
    }]
  }]
}
