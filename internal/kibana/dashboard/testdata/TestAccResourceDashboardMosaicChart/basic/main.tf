variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title                  = var.dashboard_title
  description            = "Dashboard with Mosaic Panel"
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
      title                 = "Sample Mosaic"
      description           = "Test mosaic visualization"
      ignore_global_filters = false
      sampling              = 1

      legend = {
        nested               = false
        size                 = "auto"
        truncate_after_lines = 3
        visible              = "show"
      }

      value_display = {
        mode             = "percentage"
        percent_decimals = 2
      }

      filters = [
        {
          language = "kuery"
          query    = "host.name:*"
        }
      ]

      standard = {
        dataset = jsonencode({
          type = "dataView"
          id   = "metrics-*"
        })

        query = {
          language = "kuery"
          query    = ""
        }

        group_by = [
          {
            config = jsonencode({
              operation = "terms"
              fields    = ["host.name"]
              size      = 5
            })
          }
        ]

        group_breakdown_by = [
          {
            config = jsonencode({
              operation = "terms"
              fields    = ["service.name"]
              size      = 5
            })
          }
        ]

        metrics = [
          {
            config = jsonencode({
              operation     = "count"
              empty_as_null = false
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
