resource "elasticstack_kibana_stream" "nginx" {
  name        = "logs.nginx"
  space_id    = "default"
  description = "Nginx access log stream"

  wired_config {
    # Define explicit field type mappings
    fields_json = jsonencode({
      "host.name"                 = { type = "keyword" }
      "http.response.status_code" = { type = "long" }
      "http.response.bytes"       = { type = "long" }
      "url.path"                  = { type = "keyword" }
    })

    # Route documents matching the condition to a child stream
    routing_json = jsonencode([
      {
        destination = "logs.nginx.errors"
        status      = "enabled"
        where = {
          field = "http.response.status_code"
          gte   = 400
        }
      }
    ])

    # Add a Grok processing step to parse the raw log message
    processing_steps_json = jsonencode([
      {
        grok = {
          field    = "message"
          patterns = ["%%{COMBINEDAPACHELOG}"]
        }
      }
    ])

    # Retain data for 30 days using DSL lifecycle
    lifecycle_json = jsonencode({
      dsl = {
        data_retention = "30d"
      }
    })

    index_number_of_shards   = 1
    index_number_of_replicas = 1
    index_refresh_interval   = "5s"
  }

  # Attach ES|QL queries for anomaly detection
  queries = [
    {
      id             = "high-error-rate"
      title          = "High error rate"
      description    = "Detects elevated 5xx error rates by host"
      esql           = "FROM logs.nginx | WHERE http.response.status_code >= 500 | STATS count = COUNT() BY host.name"
      severity_score = 70
    }
  ]
}
