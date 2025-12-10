provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_security_role" "test" {
  name    = var.role_name
  cluster = []

  indices {
    names      = ["cluster-*", "logs-*", "metrics-*-*", "service-*", "synthetics-*-*", "traces-*-*"]
    privileges = ["read", "view_index_metadata", "read_cross_cluster", "monitor"]
    field_security {
      grant = ["*"]
    }
  }

  run_as = []

  metadata = jsonencode({})
}
