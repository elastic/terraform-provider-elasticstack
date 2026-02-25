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
    language = "kuery"
    text     = ""
  }

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
            dataset_json = jsonencode({
              type = "dataView"
              id   = "metrics-*"
            })
            ignore_global_filters = false
            sampling              = 1
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
        language = "kuery"
        query    = ""
      }
      filters = [
        {
          query = "log.level:error"
        }
      ]
    }
  }]
}
