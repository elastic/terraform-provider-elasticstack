variable "dashboard_title" {
  type = string
}

provider "elasticstack" {
  kibana {}
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
    json     = jsonencode({ match_all = {} })
  }
}
