variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title = var.dashboard_title

  refresh_interval = {
    pause = false
    value = 30000
  }
}
