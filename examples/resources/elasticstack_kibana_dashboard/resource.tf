resource "elasticstack_kibana_dashboard" "my_dashboard" {
  title       = "My Dashboard"
  description = "A dashboard showing key metrics"

  # Time range
  time_range = {
    from = "now-15m"
    to   = "now"
  }

  # Refresh settings
  refresh_interval = {
    pause = false
    value = 60000 # 60 seconds
  }

  # Query settings with text-based query (KQL or Lucene)
  query = {
    language = "kuery"
    text     = "status:success"
  }

  # Optional tags
  tags = ["production", "monitoring"]
}

# Example with JSON query
resource "elasticstack_kibana_dashboard" "my_dashboard_json" {
  title       = "My Dashboard with JSON Query"
  description = "A dashboard with a structured query"

  # Time range
  time_range = {
    from = "now-15m"
    to   = "now"
  }

  # Refresh settings
  refresh_interval = {
    pause = false
    value = 60000 # 60 seconds
  }

  # Query settings with JSON query object
  query = {
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

