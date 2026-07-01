variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title       = var.dashboard_title
  description = "Dashboard with conflicting typed panel config blocks"

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
    type = "ml_anomaly_swimlane"
    grid = {
      x = 0
      y = 0
      w = 24
      h = 8
    }
    ml_anomaly_swimlane_config = {
      swimlane_type = "overall"
      job_ids       = ["fake-job-alpha"]
    }
    slo_burn_rate_config = {
      slo_id   = "test-slo-id"
      duration = "72h"
    }
  }]
}
