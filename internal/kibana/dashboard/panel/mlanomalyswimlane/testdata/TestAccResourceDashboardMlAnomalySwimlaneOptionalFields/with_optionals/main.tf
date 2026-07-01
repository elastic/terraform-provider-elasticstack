variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title       = var.dashboard_title
  description = "Dashboard with ML anomaly swim lane panel (optional fields)"

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
      swimlane_type = "viewBy"
      job_ids       = ["fake-job-alpha", "fake-job-beta"]
      view_by       = "host.name"
      per_page      = 10
      title         = "Anomaly Swim Lane"
      description   = "View-by swim lane panel"
      hide_title    = true
      hide_border   = false
      time_range = {
        from = "now-7d"
        to   = "now"
        mode = "relative"
      }
    }
  }]
}
