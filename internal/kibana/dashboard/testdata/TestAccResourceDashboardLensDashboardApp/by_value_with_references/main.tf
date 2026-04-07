variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title       = var.dashboard_title
  description = "Dashboard with lens-dashboard-app panel (by-value, with references)"

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
    type = "lens-dashboard-app"
    grid = {
      x = 0
      y = 0
      w = 24
      h = 15
    }
    lens_dashboard_app_config = {
      by_value = {
        attributes_json = jsonencode({
          visualizationType = "lnsXY"
          title             = "Test XY Chart"
        })
        references_json = jsonencode([{
          id   = "test-data-view-id"
          name = "indexpattern-datasource-layer-test"
          type = "index-pattern"
        }])
      }
    }
  }]
}
