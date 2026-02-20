variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title                  = var.dashboard_title
  description            = "Dashboard with XY Chart Panel"
  time_from              = "now-15m"
  time_to                = "now"
  refresh_interval_pause = true
  refresh_interval_value = 0
  query_language         = "kuery"
  query_text             = ""

  panels = [{
    type = "lens"
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
        left = {
          scale = "linear"
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
            dataset = jsonencode({
              type  = "esql"
              query = "FROM metrics-* | KEEP @timestamp, host.name, system.cpu.user.pct | LIMIT 10"
            })
            x = jsonencode({
              operation = "value"
              column    = "@timestamp"
            })
            breakdown_by = jsonencode({
              operation   = "value"
              column      = "host.name"
              collapse_by = "avg"
              color = {
                mode    = "categorical"
                palette = "default"
                mapping = [
                  {
                    color = {
                      type  = "colorCode"
                      value = "#54B399"
                    }
                    values = ["host-a"]
                  }
                ]
                unassignedColor = {
                  type  = "colorCode"
                  value = "#D3DAE6"
                }
              }
            })
            y = [
              {
                config = jsonencode({
                  operation = "value"
                  column    = "system.cpu.user.pct"
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
        language = "kuery"
        query    = ""
      }
    }
  }]
}
