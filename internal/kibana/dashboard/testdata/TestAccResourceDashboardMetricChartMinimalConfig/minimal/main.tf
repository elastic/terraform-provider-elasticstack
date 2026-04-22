# Regression config for https://github.com/elastic/terraform-provider-elasticstack/issues/2355
#
# Intentionally omits optional metric_chart_config attributes that have Kibana API defaults:
#   - ignore_global_filters        (default: false)
#   - sampling                     (default: 1)
#   - query.language               (default: "kql")
#   - metrics[].config_json fields: empty_as_null (default: false), color (default: {"type":"auto"}),
#                                   format.decimals (default: 2), format.compact (default: false)
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
    metric_chart_config = {
      data_source_json = jsonencode({ type = "data_view_spec", index_pattern = "metrics-*" })
      query            = { expression = "" }
      metrics = [{
        # type, operation and format.type are required by the API.
        # Intentionally omitting optional fields with API defaults:
        #   empty_as_null (default: false), color (default: {"type":"auto"}),
        #   format.decimals (default: 2), format.compact (default: false).
        config_json = jsonencode({
          type      = "primary"
          operation = "count"
          format    = { type = "number" }
        })
      }]
    }
  }]
}
