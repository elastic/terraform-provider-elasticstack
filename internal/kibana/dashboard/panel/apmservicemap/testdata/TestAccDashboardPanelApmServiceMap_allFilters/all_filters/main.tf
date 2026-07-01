variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title       = var.dashboard_title
  description = "Dashboard with APM service map panel (all filters)"

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
      alert_status_filter = ["active", "delayed"]
      anomaly_severity_filter = ["major", "critical"]
      connection_filter = ["connected", "orphaned"]
      slo_status_filter = ["healthy", "noData"]
    }
  }]
}
