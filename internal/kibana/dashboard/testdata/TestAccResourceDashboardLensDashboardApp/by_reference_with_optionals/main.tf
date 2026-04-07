variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title       = var.dashboard_title
  description = "Dashboard with lens-dashboard-app panel (by-reference, with optional fields)"

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
      by_reference = {
        saved_object_id = "test-saved-object-id-002"
      }
      title       = "My Shared Visualization"
      description = "A shared Lens visualization"
      hide_title  = true
      hide_border = false
    }
  }]
}
