variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  space_id    = "default"
  title       = var.dashboard_title
  description = "Test for issue #1790"
  tags        = ["test"]
  time_range = {
    from = "now-7d"
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
    type = "vis"
    grid = {
      h = 15
      w = 24
      x = 0
      y = 0
    }
    id = "panel-1"
    config_json = jsonencode({
      type        = "legacy_metric"
      title       = ""
      description = ""
      filters     = []
      query = {
        language   = "kql"
        expression = ""
      }
      data_source = {
        type          = "data_view_spec"
        index_pattern = "metrics-*"
        time_field    = "@timestamp"
      }
      metric = {
        operation     = "count"
        empty_as_null = false
        format = {
          type     = "number"
          decimals = 2
          compact  = false
        }
      }
      sampling              = 1
      ignore_global_filters = false
      time_range            = { from = "now-15m", to = "now" }
    })
  }]
}
