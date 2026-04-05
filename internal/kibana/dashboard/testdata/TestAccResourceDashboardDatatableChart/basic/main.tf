variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title                  = var.dashboard_title
  description            = "Dashboard with Datatable Panel"
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
    datatable_config = {
      no_esql = {
        title       = "Sample Datatable"
        description = "Test datatable visualization"
        dataset_json = jsonencode({
          type  = "index"
          index = "metrics-*"

          time_field = "@timestamp"
        })
        density = {
          mode = "compact"
        }
        query = {
          language   = "kql"
          expression = ""
        }
        metrics = [
          {
            config_json = jsonencode({
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
        ignore_global_filters = false
        sampling              = 1
        paging                = 10
      }
    }
  }]
}
