variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title       = var.dashboard_title
  description = "Dashboard with Range Slider Control Panel (with config)"

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
      ignore_validations = false
      value              = ["100", "500"]
      step               = 10
    }
  }]
}
