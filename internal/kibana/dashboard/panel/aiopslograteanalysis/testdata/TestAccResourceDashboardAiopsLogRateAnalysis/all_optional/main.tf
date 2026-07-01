variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title       = var.dashboard_title
  description = "Dashboard with an AIOps log rate analysis panel (all optional fields)"

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
    type = "aiops_log_rate_analysis"
    grid = {
      x = 0
      y = 0
      w = 24
      h = 15
    }
    aiops_log_rate_analysis_config = {
      data_view_id = "logs-*"
      title        = "Log spikes"
      description  = "Log rate analysis panel"
      hide_title   = true
      hide_border  = false
      time_range = {
        from = "now-30m"
        to   = "now"
        mode = "relative"
      }
    }
  }]
}
