resource "elasticstack_kibana_dashboard" "my_dashboard" {
  title       = "My Dashboard"
  description = "A dashboard showing key metrics"

  time_range {
    from = "now-15m"
    to   = "now"
  }

  refresh_interval {
    pause = false
    value = 60000 # 60 seconds
  }

  query {
    language = "kuery"
    text     = "status:success"
  }

  # Optional tags
  tags = ["production", "monitoring"]
}

# Example with JSON query (mutually exclusive with query.text)
resource "elasticstack_kibana_dashboard" "my_dashboard_json" {
  title       = "My Dashboard with JSON Query"
  description = "A dashboard with a structured query"

  time_range {
    from = "now-15m"
    to   = "now"
  }

  refresh_interval {
    pause = false
    value = 60000 # 60 seconds
  }

  query {
    language = "kuery"
    json = jsonencode({
      bool = {
        must = [
          {
            match = {
              status = "success"
            }
          }
        ]
      }
    })
  }

  # Optional tags
  tags = ["production", "monitoring"]
}
