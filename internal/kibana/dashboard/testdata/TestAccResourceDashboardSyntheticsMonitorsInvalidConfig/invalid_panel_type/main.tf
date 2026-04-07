variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title = var.dashboard_title

  panels = [{
    type = "lens"
    grid = {
      x = 0
      y = 0
      w = 24
      h = 15
    }
    synthetics_monitors_config = {
      filters = {}
    }
  }]
}
