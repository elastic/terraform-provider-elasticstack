# Regression config for https://github.com/elastic/terraform-provider-elasticstack/issues/2355
#
# Intentionally omits optional xy_chart_config attributes that have Kibana API defaults:
#   - axis.x.title.visible (default: true)
#   - axis.y.title.visible (default: true)
#   - axis.y.scale         (default: "linear")
#   - legend.visibility    (default: "visible")
#   - legend.inside        (default: false)
#   - legend.position      (default: "right")
#   - decorations.fill_opacity                  (default: 0.3)
#   - query.language                            (default: "kql")
#   - layers[].data_layer.ignore_global_filters (default: false)
#   - layers[].data_layer.sampling              (default: 1)
#
# If the issue is not fixed, applying this config will fail with
# "Provider produced inconsistent result after apply".

variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title            = var.dashboard_title
  time_range       = { from = "now-15m", to = "now" }
  refresh_interval = { pause = true, value = 0 }
  query            = { language = "kql", text = "" }

  panels = [{
    type = "vis"
    grid = { x = 0, y = 0, w = 24, h = 15 }
    viz_config = {
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
