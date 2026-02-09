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
          ticks             = true
          grid              = true
          label_orientation = "horizontal"
          extent = jsonencode({
            type  = "custom"
            start = 0
            end   = 100
          })
        }
        right = {
          scale = "sqrt"
          title = {
            value   = "Rate"
            visible = true
          }
          ticks             = false
          grid              = false
          label_orientation = "vertical"
          extent = jsonencode({
            type = "focus"
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
          extent = jsonencode({
            type             = "full"
            integer_rounding = true
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
            dataset = jsonencode({
              type = "dataView"
              id   = "metrics-*"
            })
            y = [
              {
                config = jsonencode({
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
    }
  }]
}
