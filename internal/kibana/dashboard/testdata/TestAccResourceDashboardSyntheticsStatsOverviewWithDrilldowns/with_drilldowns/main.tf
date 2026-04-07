variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title       = var.dashboard_title
  description = "Dashboard with Synthetics Stats Overview Panel (drilldowns)"

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
    type = "synthetics_stats_overview"
    grid = {
      x = 0
      y = 0
      w = 24
      h = 6
    }
    synthetics_stats_overview_config = {
      drilldowns = [
        {
          url     = "https://example.com/{{context.panel.title}}"
          label   = "View details"
          trigger = "on_open_panel_menu"
          type    = "url_drilldown"
        }
      ]
    }
  }]
}
