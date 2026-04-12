variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title       = var.dashboard_title
  description = "Dashboard with SLO Single Overview Panel"

  time_range = {
    from = "now-15m"
    to   = "now"
  }
  refresh_interval = {
    pause = true
    value = 0
  }
  query = {
    language = "kql"
    text     = ""
  }
  panels = [{
    type = "slo_overview"
    grid = {
      x = 0
      y = 0
      w = 24
      h = 8
    }
    slo_overview_config = {
      single = {
        slo_id          = "my-slo-id"
        slo_instance_id = "instance-1"
        title           = "My SLO Overview"
        description     = "Displays the status of my SLO"
      }
    }
  }]
}
