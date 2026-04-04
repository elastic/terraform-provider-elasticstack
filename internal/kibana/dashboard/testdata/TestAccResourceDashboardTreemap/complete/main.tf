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
  query_language         = "kql"
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

      dataset_json = jsonencode({
        type = "index"
        index = "metrics-*"

        time_field = "@timestamp"
      })

      query = {
        language = "kql"
        expression    = "service.name:*"
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

      group_by_json = jsonencode([
        {
          operation = "terms"
          color = {
            mode    = "categorical"
            palette = "default"
            mapping = []
            unassigned = {
              type  = "color_code"
              value = "#D3DAE6"
            }
          }
          fields = ["service.name"]
          rank_by = {
            direction = "desc"
            metric    = 0
            type      = "column"
          }
          limit = 5
        }
      ])

      metrics_json = jsonencode([
        {
          operation = "count"
        }
      ])

      legend = {
        nested               = false
        size                 = "s"
        visible              = "visible"
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
