variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title       = var.dashboard_title
  description = "Dashboard with datatable presentation acceptance"
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
    viz_config = {
      by_value = {
        datatable_config = {
          no_esql = {
            title       = "Datatable presentation acc"
            description = "Cross-cutting lens presentation acceptance"
            time_range = {
              from = "now-30d"
              to   = "now-1d"
            }
            hide_title = true
            drilldowns = [
              {
                url_drilldown = {
                  url     = "https://example.com/{{event.field}}"
                  label   = "Open URL"
                  trigger = "on_click_value"
                }
              }
            ]
            data_source_json = jsonencode({
              type          = "data_view_spec"
              index_pattern = "metrics-*"
              time_field    = "@timestamp"
            })
            styling = {
              density = {
                mode = "compact"
              }
              paging = 10
            }
            query = {
              language   = "kql"
              expression = ""
            }
            metrics = [
              {
                config_json = jsonencode({
                  operation     = "count"
                  empty_as_null = false
                  format = {
                    type     = "number"
                    compact  = false
                    decimals = 2
                  }
                })
              }
            ]
            ignore_global_filters = false
            sampling              = 1
          }
        }
      }
    }
  }]
}
