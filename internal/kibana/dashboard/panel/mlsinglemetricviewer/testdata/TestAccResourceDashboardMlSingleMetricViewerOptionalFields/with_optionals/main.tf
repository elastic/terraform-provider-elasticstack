variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title       = var.dashboard_title
  description = "Dashboard with ML single metric viewer panel (optional fields)"

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
    type = "ml_single_metric_viewer"
    grid = {
      x = 0
      y = 0
      w = 24
      h = 8
    }
    ml_single_metric_viewer_config = {
      job_ids     = ["fake-job-alpha"]
      title       = "Single Metric Viewer"
      description = "SMV panel"
      hide_title  = true
      hide_border = false
      time_range = {
        from = "now-7d"
        to   = "now"
        mode = "relative"
      }
    }
  }]
}
