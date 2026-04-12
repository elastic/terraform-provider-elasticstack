variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title       = var.dashboard_title
  description = "Dashboard with XY Chart Panel"
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
    xy_chart_config = {
      title       = "Sample XY Chart"
      description = "Test XY chart visualization"
      axis = {
        y = {
          scale = "linear"
          domain_json = jsonencode({
            type = "fit"
          })
          title = {
            value   = "Count"
            visible = true
          }
        }
        x = {
          title = {
            value   = "Timestamp"
            visible = true
          }
        }
      }
      decorations = {
        fill_opacity = 0.3
      }
      fitting = {
        type = "none"
      }
      layers = [
        {
          type = "line"
          data_layer = {
            ignore_global_filters = false
            sampling              = 1
            data_source_json = jsonencode({
              type  = "esql"
              query = "FROM metrics-* | KEEP @timestamp, host.name, system.cpu.user.pct | LIMIT 10"
            })
            x_json = jsonencode({
              column = "@timestamp"
              format = {
                type = "number"
              }
            })
            breakdown_by_json = jsonencode({
              column      = "host.name"
              collapse_by = "avg"
              format = {
                type = "number"
              }
              color = {
                mode    = "categorical"
                palette = "default"
                mapping = [
                  {
                    color = {
                      type  = "color_code"
                      value = "#54B399"
                    }
                    values = ["host-a"]
                  }
                ]
                unassigned = {
                  type  = "color_code"
                  value = "#D3DAE6"
                }
              }
            })
            y = [
              {
                config_json = jsonencode({
                  column = "system.cpu.user.pct"
                  format = {
                    type = "number"
                  }
                  color = {
                    type  = "static"
                    color = "#54B399"
                  }
                })
              }
            ]
          }
        }
      ]
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
  }]
}
