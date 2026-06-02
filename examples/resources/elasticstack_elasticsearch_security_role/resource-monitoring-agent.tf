provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_security_role" "monitoring_agent" {
  name = "monitoring_agent"

  cluster = ["monitor", "manage_index_templates"]

  # Keep restricted index access disabled unless Elasticsearch system indices are
  # absolutely required. Enabling it can grant broad access to sensitive data.
  indices {
    names                    = [".monitoring-*", "metricbeat-*"]
    privileges               = ["write", "create_index"]
    allow_restricted_indices = false
  }
}
