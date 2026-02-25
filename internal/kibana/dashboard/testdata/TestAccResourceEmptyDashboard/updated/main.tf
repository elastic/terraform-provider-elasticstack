variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title       = var.dashboard_title
  description = "Updated dashboard description"

  time_range = {
    from = "now-30m"
    to   = "now"
  }

  refresh_interval = {
    pause = false
    value = 30000
  }

  query = {
    language = "kuery"
    text     = ""
  }
}
