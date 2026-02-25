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

  time_range = {
    from = "now-15m"
    to   = "now"
  }

  refresh_interval = {
    pause = true
    value = 60000
  }

  query = {
    language = "kuery"
    text     = ""
  }
}
