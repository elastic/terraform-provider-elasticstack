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
          title = {
            value   = "Count"
            visible = true
          }
          ticks             = true
          grid              = true
          label_orientation = "horizontal"
          domain_json = jsonencode({
            type = "custom"
            min  = 0
            max  = 100
          })
        }
        secondary_y = {
          scale = "sqrt"
          title = {
            value   = "Rate"
            visible = true
          }
          ticks             = false
          grid              = false
          label_orientation = "vertical"
          domain_json = jsonencode({
            type = "fit"
          })
        }
        x = {
          title = {
            value   = "Timestamp"
            visible = true
          }
          ticks             = true
          grid              = true
          label_orientation = "angled"
          domain_json = jsonencode({
            type = "fit"
          })
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
              type          = "data_view_spec"
              index_pattern = "metrics-*"
            })
            y = [
              {
                config_json = jsonencode({
                  operation     = "count"
                  empty_as_null = true
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
