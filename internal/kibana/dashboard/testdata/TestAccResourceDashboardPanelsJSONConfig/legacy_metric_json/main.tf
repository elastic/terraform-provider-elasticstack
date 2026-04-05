variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  space_id               = "default"
  title                  = var.dashboard_title
  description            = "Test for issue #1790"
  tags                   = ["test"]
  time_from              = "now-7d"
  time_to                = "now"
  refresh_interval_pause = true
  refresh_interval_value = 0
  query_language         = "kql"
  query_text             = ""

  panels = [{
    type = "lens"
    grid = {
      h = 15
      w = 24
      x = 0
      y = 0
    }
    uid = "panel-1"
    config_json = jsonencode({
      attributes = {
        dataset = {
          type  = "index"
          index = "metrics-*"

          time_field = "@timestamp"
        }
        description           = ""
        filters               = []
        ignore_global_filters = false
        metric = {
          operation     = "count"
          empty_as_null = false
          format = {
            type     = "number"
            decimals = 2
            compact  = false
          }
        }
        sampling = 1
        title    = ""
        type     = "legacy_metric"
      }
      time_range = { from = "now-15m", to = "now" }
    })
  }]
}
