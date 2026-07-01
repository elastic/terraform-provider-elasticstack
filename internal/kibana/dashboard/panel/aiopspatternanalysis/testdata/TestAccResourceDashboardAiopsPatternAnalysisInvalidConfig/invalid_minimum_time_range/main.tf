variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title       = var.dashboard_title
  description = "Dashboard with an AIOps pattern analysis panel (invalid minimum_time_range)"

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
    type = "aiops_pattern_analysis"
    grid = {
      x = 0
      y = 0
      w = 24
      h = 15
    }
    aiops_pattern_analysis_config = {
      data_view_id       = "logs-*"
      field_name         = "message"
      minimum_time_range = "2_weeks"
    }
  }]
}
