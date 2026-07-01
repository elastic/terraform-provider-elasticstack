variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title       = var.dashboard_title
  description = "Dashboard with ML single metric viewer selected_entities"

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
      job_ids = ["fake-job-alpha"]
      selected_entities = {
        airline = {
          string_value = "AAL"
        }
        region_code = {
          numeric_value = 4
        }
      }
    }
  }]
}
