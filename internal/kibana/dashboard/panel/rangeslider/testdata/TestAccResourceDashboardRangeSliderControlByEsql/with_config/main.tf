variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title       = var.dashboard_title
  description = "Dashboard with Range Slider Control Panel (by_esql)"

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
      by_esql = {
        esql_query         = "FROM logs-* | STATS min = MIN(bytes), max = MAX(bytes)"
        values_source      = "esql_query"
        title              = "Bytes Range"
        use_global_filters = true
        ignore_validations = false
        step               = 10
      }
    }
  }]
}
