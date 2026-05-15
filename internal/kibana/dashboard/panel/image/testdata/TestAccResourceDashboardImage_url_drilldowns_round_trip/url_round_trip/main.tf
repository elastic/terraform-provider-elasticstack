variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "target" {
  title       = "${var.dashboard_title} target"
  description = "Drilldown target for image panel acceptance"

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
}

resource "elasticstack_kibana_dashboard" "test" {
  title       = var.dashboard_title
  description = "Dashboard with typed image panel (URL src, mixed drilldowns)"

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
    type = "image"
    grid = {
      x = 0
      y = 0
      w = 24
      h = 10
    }
    image_config = {
      src = {
        url = {
          url = "https://example.com/logo.png"
        }
      }
      alt_text         = "Logo"
      background_color = "#111111"
      title            = "Image panel title"
      description      = "Image panel description"
      hide_title       = false
      hide_border      = true
      drilldowns = [
        {
          dashboard_drilldown = {
            dashboard_id = elasticstack_kibana_dashboard.target.dashboard_id
            label        = "Open dashboard"
            trigger      = "on_click_image"
          }
        },
        {
          url_drilldown = {
            url     = "https://example.com/details/{{context.panel.title}}"
            label   = "External link"
            trigger = "on_click_image"
          }
        },
      ]
    }
  }]

  depends_on = [elasticstack_kibana_dashboard.target]
}
