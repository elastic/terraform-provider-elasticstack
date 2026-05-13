variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title            = var.dashboard_title
  time_range       = { from = "now-7d", to = "now" }
  refresh_interval = { pause = true, value = 0 }
  query            = { language = "kql", text = "" }

  panels = [{
    type = "vis"
    grid = { x = 0, y = 0, w = 24, h = 15 }
    vis_config = {
      by_value = {
        xy_chart_config = {
          time_range = null
          axis = {
            y = { domain_json = jsonencode({ type = "fit" }) }
          }
          legend      = {}
          fitting     = { type = "none" }
          decorations = {}
          query       = { expression = "" }
          layers = [{
            type = "area"
            data_layer = {
              data_source_json = jsonencode({ type = "data_view_spec", index_pattern = "logs-*" })
              y = [{
                config_json = jsonencode({ operation = "count", empty_as_null = true })
              }]
            }
          }]
        }
      }
    }
  }]
}
