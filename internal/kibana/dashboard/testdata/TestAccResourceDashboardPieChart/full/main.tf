variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title                  = var.dashboard_title
  description            = "Dashboard with Pie Chart Panel (Full)"
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
    pie_chart_config = {
      title          = "Full Pie Chart"
      description    = "Full pie chart visualization"
      donut_hole     = "large"
      label_position = "outside"
      dataset = jsonencode({
        type = "dataView"
        id   = "metrics-*"
      })
      query = {
        language = "kuery"
        query    = ""
      }
      legend = jsonencode({
        visible = "show"
        size    = "auto"
      })
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
              unassignedColor = {
                type  = "colorCode"
                value = "#555555"
              }
            }
          })
        }
      ]
      ignore_global_filters = false // Default value
      sampling              = 1     // Default value
      filters = [
        {
          query    = "response:200"
          language = "kuery"
        }
      ]
    }
  }]
}
