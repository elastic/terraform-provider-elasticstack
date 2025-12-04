variable "space_name" {
  type = string
}

variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_space" "test" {
  space_id = var.space_name
  name     = var.space_name
}

resource "elasticstack_kibana_dashboard" "test_space" {
  space_id    = elasticstack_kibana_space.test.space_id
  title       = var.dashboard_title
  description = "Test dashboard in custom space"

  time_from = "now-15m"
  time_to   = "now"

  refresh_interval_pause = true
  refresh_interval_value = 60000

  query_language = "kuery"
  query_text     = ""
}
