provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_security_api_key" "data_analyst_logs_only" {
  name = "data-analyst-logs-only"

  role_descriptors = jsonencode({
    logs_only = {
      cluster = ["monitor"]
      indices = [
        {
          names      = ["logs-myapp-*"]
          privileges = ["read", "view_index_metadata"]
        }
      ]
    }
  })
}
