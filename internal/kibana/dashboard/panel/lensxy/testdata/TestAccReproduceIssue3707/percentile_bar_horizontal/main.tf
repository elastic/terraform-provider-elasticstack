# Regression config for https://github.com/elastic/terraform-provider-elasticstack/issues/3707
#
# Using operation = "percentile" in the y config_json for a bar_horizontal layer
# previously failed: the provider unconditionally injected empty_as_null=false, which the
# Kibana percentile metric schema rejects with HTTP 400. The config intentionally omits
# empty_as_null so the provider's default injection (now gated by operation) is exercised.

variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "repro_3707" {
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
                value   = "Responses"
                visible = true
              }
              domain_json = jsonencode({ type = "fit" })
            }
            y2 = {
              title = {
                value   = "p95"
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
                  field      = "http.response.duration"
                  operation  = "percentile"
                  percentile = 95
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
