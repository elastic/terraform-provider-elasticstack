variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title       = var.dashboard_title
  description = "Dashboard with pie chart presentation acceptance"
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
    vis_config = {
      by_value = {
        pie_chart_config = {
          title       = "Pie presentation acc"
          description = "Cross-cutting lens presentation acceptance"
          time_range = {
            from = "now-30d"
            to   = "now-1d"
          }
          hide_title = true
          drilldowns = [
            {
              discover_drilldown = {
                label = "Open Discover"
              }
            }
          ]
          donut_hole     = "s"
          label_position = "inside"
          data_source_json = jsonencode({
            type          = "data_view_spec"
            index_pattern = "metrics-*"
            time_field    = "@timestamp"
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
      }
    }
  }]
}
