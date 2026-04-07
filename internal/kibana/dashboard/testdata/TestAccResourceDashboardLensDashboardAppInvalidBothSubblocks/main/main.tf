variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title = var.dashboard_title

  time_range = {
    from = "now-15m"
    to   = "now"
  }
  panels = [{
    type = "lens-dashboard-app"
    grid = {
      x = 0
      y = 0
    }
    lens_dashboard_app_config = {
      by_reference = {
        saved_object_id = "abc"
      }
      by_value = {
        attributes_json = "{}"
      }
    }
  }]
}
