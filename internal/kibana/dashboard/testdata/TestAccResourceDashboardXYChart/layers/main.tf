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
            ignore_global_filters = true
            sampling              = 0.5
            dataset = jsonencode({
              type = "dataView"
              id   = "metrics-*"
            })
            x = jsonencode({
              operation               = "date_histogram"
              field                   = "@timestamp"
              drop_partial_intervals  = false
              include_empty_rows      = true
              suggested_interval      = "auto"
              use_original_time_range = false
            })
            breakdown_by = jsonencode({
              operation = "terms"
              fields    = ["host.name"]
              size      = 5
              rank_by = {
                direction = "desc"
                metric    = 0
                type      = "column"
              }
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
