variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title                  = var.dashboard_title
  description            = "Dashboard with Mosaic Panel (ES|QL)"
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

    mosaic_config = {
      title                 = "Sample Mosaic ESQL"
      description           = "Test mosaic visualization (ES|QL)"
      ignore_global_filters = false
      sampling              = 1

      legend = {
        size = "auto"
      }

      value_display = {
        mode = "absolute"
      }

      filters = [
        {
          language = "kuery"
          query    = "host.name:*"
        }
      ]

      esql = {
        dataset = jsonencode({
          type  = "esql"
          query = "FROM metrics-* | KEEP host.name, service.name, system.cpu.user.pct | LIMIT 10"
        })

        group_by = [
          {
            config = jsonencode({
              operation   = "value"
              column      = "host.name"
              collapse_by = "avg"
              color = {
                mode    = "gradient"
                palette = "default"
                unassignedColor = {
                  type  = "colorCode"
                  value = "#D3DAE6"
                }
              }
            })
          }
        ]

        group_breakdown_by = [
          {
            config = jsonencode({
              operation   = "value"
              column      = "service.name"
              collapse_by = "avg"
              color = {
                mode    = "gradient"
                palette = "default"
                unassignedColor = {
                  type  = "colorCode"
                  value = "#D3DAE6"
                }
              }
            })
          }
        ]

        metrics = [
          {
            config = jsonencode({
              operation = "value"
              column    = "system.cpu.user.pct"
              format = {
                type     = "number"
                compact  = false
                decimals = 2
              }
            })
          }
        ]
      }
    }
  }]
}
