variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title                  = var.dashboard_title
  description            = "Dashboard with Heatmap Panel"
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
    type = "lens"
    grid = {
      x = 0
      y = 0
      w = 24
      h = 15
    }
    heatmap_config = {
      title       = "Sample Heatmap"
      description = "Test heatmap visualization"
      dataset_json = jsonencode({
        type  = "index"
        index = "metrics-*"

        time_field = "@timestamp"
      })
      query = {
        language   = "kql"
        expression = ""
      }
      metric_json = jsonencode({
        operation = "count"
      })
      x_axis_json = jsonencode({
        operation = "filters"
        filters = [
          {
            label = "All"
            filter = {
              expression = "*"
              language   = "kql"
            }
          }
        ]
      })
      y_axis_json = jsonencode({
        operation = "filters"
        filters = [
          {
            label = "All"
            filter = {
              expression = "*"
              language   = "kql"
            }
          }
        ]
      })
      axes = {
        x = {
          labels = {
            orientation = "horizontal"
            visible     = true
          }
          title = {
            value   = "X Axis"
            visible = true
          }
        }
        y = {
          labels = {
            visible = true
          }
          title = {
            value   = "Y Axis"
            visible = true
          }
        }
      }
      cells = {
        labels = {
          visible = true
        }
      }
      legend = {
        visibility           = "visible"
        size                 = "m"
        truncate_after_lines = 5
      }
      ignore_global_filters = false
      sampling              = 1
    }
  }]
}
