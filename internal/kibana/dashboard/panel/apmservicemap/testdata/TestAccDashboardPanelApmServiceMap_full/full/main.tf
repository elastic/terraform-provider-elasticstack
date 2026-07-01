variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title       = var.dashboard_title
  description = "Dashboard with APM service map panel (full configuration)"

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
      title                       = "APM Service Map"
      description                 = "Dependencies overview"
      hide_title                  = true
      hide_border                 = false
      environment                 = "production"
      service_name                = "checkout"
      service_group_id            = "group-abc"
      kuery                       = "service.name : checkout"
      map_orientation             = "vertical"
      sync_with_dashboard_filters = true
      alert_status_filter         = ["active", "recovered"]
      anomaly_severity_filter     = ["warning", "minor"]
      connection_filter           = ["connected"]
      slo_status_filter           = ["degrading", "violated"]
      time_range = {
        from = "now-7d"
        to   = "now"
        mode = "relative"
      }
    }
  }]
}
