# Regression config for https://github.com/elastic/terraform-provider-elasticstack/issues/3402
#
# A bar_stacked XY layer combined with an explicit `fitting.type = "none"` previously
# triggered "Provider produced inconsistent result after apply" because Kibana's
# read-back returned `fitting.type = ""` (empty string). The provider now treats
# the empty string as semantically null so the practitioner's "none" round-trips.

variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "repro_3402" {
  title = var.dashboard_title
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
              domain_json = jsonencode({ type = "fit" })
            }
          }
          decorations = {}
          fitting = {
            type = "none"
          }
          layers = [{
            type = "bar_stacked"
            data_layer = {
              data_source_json = jsonencode({
                type          = "data_view_spec"
                index_pattern = "metrics-*"
              })
              y = [{
                config_json = jsonencode({
                  operation     = "count"
                  empty_as_null = true
                })
              }]
            }
          }]
          legend = {}
          query  = { expression = "" }
        }
      }
    }
  }]
}
