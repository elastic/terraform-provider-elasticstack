variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title = var.dashboard_title

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
    type = "field_stats_table"
    grid = {
      x = 0
      y = 0
      w = 24
      h = 15
    }
    field_stats_table_config = {
      by_esql = {
        query              = "FROM logs-* | LIMIT 100"
        show_distributions = false
        title              = "Field statistics — logs by service"
        time_range = {
          from = "now-24h"
          to   = "now"
        }
      }
    }
  }]
}
