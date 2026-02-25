variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title       = var.dashboard_title
  description = "Dashboard with ES|QL Datatable Panel"

  time_range = {
    from = "now-15m"
    to   = "now"
  }

  refresh_interval = {
    pause = true
    value = 0
  }

  query = {
    language = "kuery"
    text     = ""
  }

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
        dataset_json = jsonencode({
          type  = "esql"
          query = "FROM metrics-* | STATS count = COUNT(*) BY TBUCKET(5m)"
        })
        density = {
          mode = "default"
        }
        metrics = [
          {
            config_json = jsonencode({
              operation = "value"
              column    = "TBUCKET(5m)"
            })
          },
          {
            config_json = jsonencode({
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