provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title                  = var.dashboard_title
  description            = "Test dashboard with access control"
  time_from              = "now-15m"
  time_to                = "now"
  refresh_interval_pause = true
  refresh_interval_value = 90000
  query_language         = "kuery"
  query_text             = ""

  access_control = {
    access_mode = "write_restricted"
    owner       = "elastic"
  }
}
