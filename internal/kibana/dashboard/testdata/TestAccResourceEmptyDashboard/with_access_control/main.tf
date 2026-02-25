provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title       = var.dashboard_title
  description = "Test dashboard with access control"

  time_range = {
    from = "now-15m"
    to   = "now"
  }

  refresh_interval = {
    pause = true
    value = 90000
  }

  query = {
    language = "kuery"
    text     = ""
  }

  access_control = {
    access_mode = "write_restricted"
    owner       = "elastic"
  }
}
