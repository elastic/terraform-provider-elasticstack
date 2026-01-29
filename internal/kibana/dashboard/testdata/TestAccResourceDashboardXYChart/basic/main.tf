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
        fill_opacity      = 0.3
        point_visibility  = true
        show_value_labels = false
      }
      fitting = {
        type = "none"
      }
      layers = jsonencode([
        {
          type    = "area"
          dataset = {}
          y = [
            {
              operation = "count"
              color     = "#68BC00"
              axis      = "left"
            }
          ]
        }
      ])
      legend = {
        visible  = true
        inside   = false
        position = "right"
      }
      query = {
        language = "kuery"
        query    = ""
      }
    }
  }]
}
