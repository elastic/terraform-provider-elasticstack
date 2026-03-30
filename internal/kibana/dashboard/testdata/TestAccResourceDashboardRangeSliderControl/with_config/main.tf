variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title       = var.dashboard_title
  description = "Dashboard with Range Slider Control Panel (with config)"

  time_from = "now-15m"
  time_to   = "now"

  refresh_interval_pause = true
  refresh_interval_value = 0

  query_language = "kuery"
  query_text     = ""

  panels = [{
    type = "range_slider_control"
    grid = {
      x = 0
      y = 0
      w = 24
      h = 4
    }
    range_slider_control_config = {
      data_view_id       = "test-data-view-id"
      field_name         = "bytes"
      title              = "Bytes Range"
      use_global_filters = true
      value              = ["100", "500"]
    }
  }]
}
