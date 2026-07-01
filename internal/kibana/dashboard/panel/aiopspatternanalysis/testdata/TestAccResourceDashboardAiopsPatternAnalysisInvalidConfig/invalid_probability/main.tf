variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title       = var.dashboard_title
  description = "Dashboard with an AIOps pattern analysis panel (out-of-range probability)"

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
      data_view_id               = "logs-*"
      field_name                 = "message"
      random_sampler_mode        = "on_manual"
      random_sampler_probability = 1.0
    }
  }]
}
