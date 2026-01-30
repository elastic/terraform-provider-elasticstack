variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title                  = var.dashboard_title
  description            = "Dashboard with Top-level Panels and Sections"
  time_from              = "now-15m"
  time_to                = "now"
  refresh_interval_pause = true
  refresh_interval_value = 0
  query_language         = "kuery"
  query_text             = ""

  panels = [{
    type = "DASHBOARD_MARKDOWN"
    grid = {
      x = 0
      y = 0
      w = 24
      h = 6
    }
    markdown_config = {
      content           = "Top-level panel one"
      title             = "Top Panel One"
      hide_panel_titles = true
    }
    }, {
    type = "DASHBOARD_MARKDOWN"
    grid = {
      x = 0
      y = 6
      w = 24
      h = 6
    }
    markdown_config = {
      content           = "Top-level panel two"
      title             = "Top Panel Two"
      hide_panel_titles = true
    }
  }]

  sections = [{
    title = "Section One"
    grid = {
      y = 12
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
        content           = "Section one - panel one"
        title             = "Section One Panel"
        hide_panel_titles = false
      }
    }]
    }, {
    title = "Section Two"
    grid = {
      y = 22
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
        content           = "Section two - panel one"
        title             = "Section Two Panel"
        hide_panel_titles = false
      }
    }]
  }]
}
