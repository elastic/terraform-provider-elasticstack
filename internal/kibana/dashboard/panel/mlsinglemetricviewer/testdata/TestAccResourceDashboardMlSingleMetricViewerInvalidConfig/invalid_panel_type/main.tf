variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title       = var.dashboard_title
  description = "Dashboard with ml_single_metric_viewer_config on markdown panel"

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
    type = "markdown"
    grid = {
      x = 0
      y = 0
      w = 24
      h = 4
    }
    ml_single_metric_viewer_config = {
      job_ids = ["fake-job-alpha"]
    }
  }]
}
