variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title       = var.dashboard_title
  description = "Dashboard with Region Map Panel"
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
    region_map_config = {
      title       = "Sample Region Map"
      description = "Test region map visualization"
      data_source_json = jsonencode({
        type          = "data_view_spec"
        index_pattern = "metrics-*"

        time_field = "@timestamp"
      })
      query = {
        language   = "kql"
        expression = ""
      }
      metric_json = jsonencode({
        operation = "count"
      })
      region_json = jsonencode({
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
      ignore_global_filters = false
      sampling              = 1
      filters = [
        {
          filter_json = jsonencode({
            type = "condition"
            condition = {
              field    = "status"
              operator = "is"
              value    = "active"
            }
          })
        }
      ]
    }
  }]
}
