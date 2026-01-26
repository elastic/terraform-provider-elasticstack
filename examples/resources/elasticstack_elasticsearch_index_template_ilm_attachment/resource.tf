# Create an ILM policy
resource "elasticstack_elasticsearch_index_lifecycle" "retention_90d" {
  name = "90-days-retention"

  hot {
    rollover {
      max_age = "7d"
    }
  }

  delete {
    min_age = "90d"
    delete {}
  }
}

# Attach the ILM policy to a Fleet-managed template
resource "elasticstack_elasticsearch_index_template_ilm_attachment" "logs_system" {
  index_template = "logs-system.syslog"
  lifecycle_name = elasticstack_elasticsearch_index_lifecycle.retention_90d.name
}
