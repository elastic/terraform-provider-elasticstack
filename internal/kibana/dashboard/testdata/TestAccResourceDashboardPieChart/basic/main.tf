variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title       = var.dashboard_title
  description = "Dashboard with Pie Chart Panel"
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
    pie_chart_config = {
      title          = "Sample Pie Chart"
      description    = "Test pie chart visualization"
      donut_hole     = "s"
      label_position = "inside"
      data_source_json = jsonencode({
        type          = "data_view_spec"
        index_pattern = "metrics-*"

        time_field = "@timestamp"
      })
      query = {
        language   = "kql"
        expression = ""
      }
      metrics = [
        {
          config = jsonencode({
            operation = "count"
            format    = { type = "number" }
          })
        }
      ]
      group_by = [
        {
          config = jsonencode({
            operation = "terms"
            fields    = ["DestCountry"]
            color = {
              mode    = "categorical"
              palette = "default"
              mapping = []
              unassigned = {
                type  = "color_code"
                value = "#555555"
              }
            }
            limit = 5
            rank_by = {
              direction    = "desc"
              metric_index = 0
              type         = "metric"
            }
          })
        }
      ]
      ignore_global_filters = false
      sampling              = 1
    }
  }]
}
