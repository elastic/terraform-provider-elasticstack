variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title       = var.dashboard_title
  description = "Dashboard with an AIOps pattern analysis panel (all optional fields)"

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
      minimum_time_range         = "1_week"
      random_sampler_mode        = "on_manual"
      random_sampler_probability = 0.01
      title                      = "Patterns"
      description                = "Pattern analysis panel"
      hide_title                 = true
      hide_border                = false
      time_range = {
        from = "now-7d"
        to   = "now"
      }
    }
  }]
}
