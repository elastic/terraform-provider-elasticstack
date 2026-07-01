variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title       = var.dashboard_title
  description = "Dashboard with an AIOps change point chart panel (required fields only)"

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
    type = "aiops_change_point_chart"
    grid = {
      x = 0
      y = 0
      w = 24
      h = 15
    }
    aiops_change_point_chart_config = {
      data_view_id = "metrics-*"
      metric_field = "system.cpu.total.pct"
    }
  }]
}
