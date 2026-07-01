variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title       = var.dashboard_title
  description = "Dashboard with ML anomaly swim lane panel (optional fields updated)"

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
      per_page      = 25
      title         = "Updated Swim Lane"
      description   = "Updated description"
      hide_title    = false
      hide_border   = true
      time_range = {
        from = "now-30d"
        to   = "now"
        mode = "absolute"
      }
    }
  }]
}
