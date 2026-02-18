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
      title       = "Complete Treemap"
      description = "Complete treemap visualization"

      dataset = jsonencode({
        type = "dataView"
        id   = "metrics-*"
      })

      query = {
        language = "kuery"
        query    = "service.name:*"
      }

      filters = [
        {
          query    = "host.os.keyword: \"linux\""
          language = "kuery"
        }
      ]

      group_by = jsonencode([
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
          fields = ["service.name"]
          size   = 5
        }
      ])

      metrics = jsonencode([
        {
          operation = "count"
        }
      ])

      label_position = "hidden"

      legend = {
        nested               = false
        size                 = "small"
        visible              = "show"
        truncate_after_lines = 10
      }

      value_display = {
        mode = "absolute"
      }

      ignore_global_filters = true
      sampling              = 0.5
    }
  }]
}
