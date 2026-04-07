variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title       = var.dashboard_title
  description = "Dashboard with Synthetics Stats Overview Panel (display settings)"

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
      title       = "Synthetics Overview"
      description = "Shows all monitor statuses"
      hide_title  = true
      hide_border = false
    }
  }]
}
