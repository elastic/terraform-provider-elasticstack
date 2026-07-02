variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title       = var.dashboard_title
  description = "Dashboard with ML anomaly charts panel (optional fields)"

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
    type = "ml_anomaly_charts"
    grid = {
      x = 0
      y = 0
      w = 24
      h = 8
    }
    ml_anomaly_charts_config = {
      job_ids            = ["fake-job-alpha"]
      max_series_to_plot = 12
      title              = "Anomaly Charts"
      description        = "ML anomaly charts panel"
      hide_title         = true
      hide_border        = false
      time_range = {
        from = "now-7d"
        to   = "now"
        mode = "relative"
      }
    }
  }]
}
