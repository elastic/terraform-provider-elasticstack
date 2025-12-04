variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title       = var.dashboard_title
  description = "Updated dashboard description"

  time_from = "now-30m"
  time_to   = "now"

  refresh_interval_pause = false
  refresh_interval_value = 30000

  query_language = "kuery"
  query_text     = ""
}
