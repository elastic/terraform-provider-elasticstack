# Companion to the issue #3707 regression: operation = "average" maps to the Kibana
# XY stats-metric schema, which (like percentile) does not accept empty_as_null. The
# config omits empty_as_null so the provider's gated default injection is exercised; a
# clean apply confirms the gate prevents the previously-broken HTTP 400.

variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title = var.dashboard_title
  time_range = {
    from = "now-1d"
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
      x = 0
      y = 0
      w = 24
      h = 15
    }
    vis_config = {
      by_value = {
        xy_chart_config = {
          axis = {
            y = {
              title = {
                value   = "Avg duration"
                visible = true
              }
              domain_json = jsonencode({ type = "fit" })
            }
            y2 = {
              title = {
                value   = "y2"
                visible = true
              }
              scale       = "linear"
              domain_json = jsonencode({ type = "fit" })
            }
          }
          decorations = {
            minimum_bar_height = 1
            show_value_labels  = false
          }
          fitting = {
            type = "none"
          }
          layers = [{
            type = "bar_horizontal"
            data_layer = {
              ignore_global_filters = false
              sampling              = 1
              data_source_json = jsonencode({
                type          = "data_view_spec"
                index_pattern = "metrics-*"
                time_field    = "@timestamp"
              })
              y = [{
                config_json = jsonencode({
                  field     = "http.response.duration"
                  operation = "average"
                  color = {
                    type = "auto"
                  }
                })
              }]
              x_json = jsonencode({
                operation = "terms"
                fields    = ["http.request.path"]
                limit     = 9
                rank_by = {
                  type         = "metric"
                  metric_index = 0
                  direction    = "desc"
                }
              })
            }
          }]
          legend = {
            visibility = "visible"
            inside     = false
            position   = "right"
          }
          query = {
            language   = "kql"
            expression = ""
          }
        }
      }
    }
  }]
}
