variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title            = var.dashboard_title
  time_range       = { from = "now-15m", to = "now" }
  refresh_interval = { pause = true, value = 0 }
  query            = { language = "kql", expression = "" }

  panels = [{
    type = "vis"
    grid = { x = 0, y = 0, w = 24, h = 15 }
    vis_config = {
      by_value = {
        xy_chart_config = {
          axis = {
            y = { domain_json = jsonencode({ type = "fit" }) }
          }
          legend      = {}
          fitting     = { type = "none" }
          decorations = {}
          query       = { expression = "" }
          layers = [{
            type = "line"
            data_layer = {
              data_source_json = jsonencode({ type = "data_view_spec", index_pattern = "logs-*" })
              y = [{
                config_json = jsonencode({
                  operation     = "count"
                  empty_as_null = true
                  format = {
                    type     = "duration"
                    from     = "s"
                    to       = "ms"
                    decimals = 0
                    compact  = true
                  }
                })
              }]
            }
          }]
        }
      }
    }
  }]
}
