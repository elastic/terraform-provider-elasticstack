variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title       = var.dashboard_title
  description = "Dashboard with Options List Control Panel (by_esql)"

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
    type = "options_list_control"
    grid = {
      x = 0
      y = 0
      w = 12
      h = 4
    }
    options_list_control_config = {
      by_esql = {
        esql_query       = "FROM logs-* | STATS BY host.name"
        values_source    = "esql_query"
        title            = "Host name"
        search_technique = "prefix"
        single_select    = true
      }
    }
  }]
}
