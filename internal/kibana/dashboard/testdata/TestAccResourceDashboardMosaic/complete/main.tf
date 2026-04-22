variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title       = var.dashboard_title
  description = "Dashboard with Mosaic Panel"
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

    mosaic_config = {
      title       = ""
      description = ""

      data_source_json = jsonencode({
        type          = "data_view_spec"
        index_pattern = "metrics-*"

        time_field = "@timestamp"
      })

      query = {
        language   = "kql"
        expression = "service.name:*"
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
          limit  = 5
          rank_by = {
            direction    = "desc"
            metric_index = 0
            type         = "metric"
          }
        }
      ])

      group_breakdown_by_json = jsonencode([
        {
          operation = "terms"
          fields    = ["host.name"]
          limit     = 5
          rank_by = {
            direction    = "desc"
            metric_index = 0
            type         = "metric"
          }
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
