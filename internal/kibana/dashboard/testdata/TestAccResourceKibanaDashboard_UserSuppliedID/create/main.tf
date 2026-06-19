variable "dashboard_title" {
  type = string
}

variable "dashboard_id" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  dashboard_id = var.dashboard_id
  title        = var.dashboard_title
  description  = "User-supplied dashboard id"

  time_range = {
    from = "now-15m"
    to   = "now"
  }
  refresh_interval = {
    pause = true
    value = 90000
  }
  query = {
    language = "kql"
    text     = "http.response.status_code:200"
  }
}
