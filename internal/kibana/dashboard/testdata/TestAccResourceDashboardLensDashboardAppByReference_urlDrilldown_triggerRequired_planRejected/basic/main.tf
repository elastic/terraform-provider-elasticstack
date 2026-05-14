variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title       = var.dashboard_title
  description = "lens-dashboard-app by_reference URL drilldown missing trigger (invalid)"
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
      h = 10
    }
    lens_dashboard_app_config = {
      by_reference = {
        ref_id = "lensRef"
        references_json = jsonencode([
          {
            id   = "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
            name = "lensRef"
            type = "lens"
          }
        ])
        time_range = {
          from = "now-7d"
          to   = "now"
          mode = "relative"
        }
        drilldowns = [
          {
            url = {
              url   = "https://example.com/{{event.field}}"
              label = "External"
            }
          }
        ]
      }
    }
  }]
}
