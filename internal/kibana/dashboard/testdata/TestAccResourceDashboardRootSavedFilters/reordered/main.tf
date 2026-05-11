variable "dashboard_title" {
  type = string
}

resource "elasticstack_kibana_dashboard" "test" {
  title       = var.dashboard_title
  description = "Dashboard root saved filters (reordered)"

  time_range = {
    from = "now-15m"
    to   = "now"
  }
  refresh_interval = {
    pause = true
    value = 90000
  }
  query = {
    language = "kql"
    text     = "http.response.status_code:200"
  }

  filters = [
    {
      filter_json = jsonencode({
        type = "condition"
        condition = {
          field    = "service.name"
          operator = "is"
          value    = "api"
        }
      })
    },
    {
      filter_json = jsonencode({
        type = "condition"
        condition = {
          field    = "host.name"
          operator = "is"
          value    = "web-01"
        }
      })
    },
  ]
}
