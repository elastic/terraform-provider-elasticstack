variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title                  = var.dashboard_title
  description            = "Dashboard with ES|QL Datatable Panel"
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
    datatable_config = {
      esql = {
        title = "count"
        dataset = jsonencode({
          type  = "esql"
          query = "FROM metrics-* | STATS count = COUNT(*) BY TBUCKET(5m)"
        })
        density = {
          mode = "default"
        }
        metrics = [
          {
            config = jsonencode({
              operation = "value"
              column    = "TBUCKET(5m)"
            })
          },
          {
            config = jsonencode({
              operation = "value"
              column    = "count"
            })
          }
        ]
        ignore_global_filters = false
        sampling              = 1
      }
    }
  }]
}