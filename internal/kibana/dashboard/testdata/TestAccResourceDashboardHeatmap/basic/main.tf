variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title                  = var.dashboard_title
  description            = "Dashboard with Heatmap Panel"
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
    heatmap_config = {
      title       = "Sample Heatmap"
      description = "Test heatmap visualization"
      dataset_json = jsonencode({
        type = "dataView"
        id   = "metrics-*"
      })
      query = {
        language = "kuery"
        query    = ""
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
              query    = "*"
              language = "kuery"
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
              query    = "*"
              language = "kuery"
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
        visible              = true
        size                 = "medium"
        position             = "right"
        truncate_after_lines = 5
      }
      ignore_global_filters = false
      sampling              = 1
    }
  }]
}
