variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title                  = var.dashboard_title
  description            = "Dashboard with Panels"
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
      h = 10
    }
    markdown_config = {
      content           = "First markdown panel"
      title             = "My Markdown Panel"
      hide_panel_titles = true
    }
    }, {
    type = "DASHBOARD_MARKDOWN"
    grid = {
      x = 0
      y = 0
      w = 24
      h = 10
    }
    markdown_config = {
      content           = "Second markdown panel"
      title             = "My Markdown Panel"
      hide_panel_titles = true
    }
  }]
}
