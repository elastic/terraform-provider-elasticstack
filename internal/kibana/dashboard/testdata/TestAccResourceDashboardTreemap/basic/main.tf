variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title                  = var.dashboard_title
  description            = "Dashboard with Treemap Panel"
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

    treemap_config = {
      title       = "Sample Treemap"
      description = "Test treemap visualization"

      dataset_json = jsonencode({
        type = "dataView"
        id   = "metrics-*"
      })

      query = {
        language = "kuery"
        query    = ""
      }

      group_by_json = jsonencode([
        {
          operation = "terms"
          color = {
            mode    = "categorical"
            palette = "default"
            mapping = []
            unassignedColor = {
              type  = "colorCode"
              value = "#D3DAE6"
            }
          }
          fields = ["host.name"]
        }
      ])

      metrics_json = jsonencode([
        {
          operation = "count"
        }
      ])


      legend = {
        nested               = true
        size                 = "medium"
        visible              = "auto"
        truncate_after_lines = 5
      }

      value_display = {
        mode             = "percentage"
        percent_decimals = 2
      }

    }
  }]
}
