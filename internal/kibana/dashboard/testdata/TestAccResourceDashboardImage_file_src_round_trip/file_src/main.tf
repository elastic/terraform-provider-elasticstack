variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title       = var.dashboard_title
  description = "Image panel with file src (placeholder file_id — may be rejected by Kibana if it validates uploads)"

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
        file = {
          file_id = "acc-test-placeholder-file-id"
        }
      }
      alt_text = "placeholder file reference"
    }
  }]
}
