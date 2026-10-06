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
    language   = "kql"
    expression = jsonencode({ match_all = {} })
  }
}
