variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title       = var.dashboard_title
  description = "Discover session panel by_value ES|QL tab"

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
    type = "discover_session"
    grid = {
      x = 0
      y = 0
      w = 24
      h = 12
    }
    discover_session_config = {
      title = "ESQL Discover"
      by_value = {
        tab = {
          esql = {
            data_source_json = jsonencode({
              type  = "esql"
              query = "FROM kibana_sample_data_logs | LIMIT 50"
            })
            row_height = "auto"
          }
        }
      }
    }
  }]
}
