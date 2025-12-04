resource "elasticstack_kibana_dashboard" "my_dashboard" {
  title       = "My Dashboard"
  description = "A dashboard showing key metrics"

  # Time range
  time_from = "now-15m"
  time_to   = "now"

  # Refresh settings
  refresh_interval_pause = false
  refresh_interval_value = 60000 # 60 seconds

  # Query settings with text-based query (KQL or Lucene)
  query_language = "kuery"
  query_text     = "status:success"

  # Optional tags
  tags = ["production", "monitoring"]
}

# Example with JSON query (mutually exclusive with query_text)
resource "elasticstack_kibana_dashboard" "my_dashboard_json" {
  title       = "My Dashboard with JSON Query"
  description = "A dashboard with a structured query"

  # Time range
  time_from = "now-15m"
  time_to   = "now"

  # Refresh settings
  refresh_interval_pause = false
  refresh_interval_value = 60000 # 60 seconds

  # Query settings with JSON query object
  query_language = "kuery"
  query_json = jsonencode({
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

  # Optional tags
  tags = ["production", "monitoring"]
}

