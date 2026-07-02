variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title       = var.dashboard_title
  description = "Dashboard with APM service map panel (combined selectors)"

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
    type = "apm_service_map"
    grid = {
      x = 0
      y = 0
      w = 24
      h = 8
    }
    apm_service_map_config = {
      environment      = "staging"
      service_name     = "checkout"
      service_group_id = "group-abc"
    }
  }]
}
