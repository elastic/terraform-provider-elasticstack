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
      title       = "Complete Heatmap"
      description = "Complete heatmap visualization"
      dataset_json = jsonencode({
        type  = "index"
        index = "metrics-*"

        time_field = "@timestamp"
      })
      query = {
        language   = "kql"
        expression = ""
      }
      filters = [
        {
          filter_json = jsonencode({
            type = "condition"
            condition = {
              field    = "host.os.keyword"
              operator = "is"
              value    = "linux"
            }
          })
        }
      ]
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
            orientation = "vertical"
            visible     = false
          }
          title = {
            value   = "Custom X Axis"
            visible = true
          }
        }
        y = {
          labels = {
            visible = false
          }
          title = {
            value   = "Custom Y Axis"
            visible = true
          }
        }
      }
      cells = {
        labels = {
          visible = false
        }
      }
      legend = {
        visibility           = "hidden"
        size                 = "s"
        truncate_after_lines = 10
      }
      ignore_global_filters = true
      sampling              = 0.5
    }
  }]
}
