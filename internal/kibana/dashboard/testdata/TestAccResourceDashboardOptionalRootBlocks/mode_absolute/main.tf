variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title = var.dashboard_title

  time_range = {
    from = "2024-01-01T00:00:00.000Z"
    to   = "2024-01-02T00:00:00.000Z"
    mode = "absolute"
  }
}
