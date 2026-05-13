variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title       = var.dashboard_title
  description = "Image panel with explicit object_fit"

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
      h = 8
    }
    image_config = {
      src = {
        url = {
          url = "https://example.com/cover.png"
        }
      }
      object_fit = "cover"
    }
  }]
}
