variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title = var.dashboard_title

  time_range = {
    from = "now-24h"
    to   = "now-1h"
  }
  refresh_interval = {
    pause = false
    value = 45000
  }
  query = {
    language = "lucene"
    text     = "host.name:web-*"
  }
}
