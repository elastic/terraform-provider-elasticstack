provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

# Install a Fleet integration - this creates Fleet-managed index templates.
# Use "system" (not "tcp") to avoid conflicting with TestAccResourceIntegration_* which use tcp.
resource "elasticstack_fleet_integration" "system" {
  name         = "system"
  version      = "1.52.2"
  force        = true
  skip_destroy = true
}

# Create an ILM policy to attach
resource "elasticstack_elasticsearch_index_lifecycle" "test" {
  name = var.policy_name

  hot {
    rollover {
      max_age = "1d"
    }
  }

  delete {
    min_age = "30d"
    delete {}
  }
}

# Attach the ILM policy to the Fleet-managed template.
# The system integration creates the "logs-system.syslog" index template.
resource "elasticstack_elasticsearch_index_template_ilm_attachment" "test" {
  depends_on = [elasticstack_fleet_integration.system]

  index_template = "logs-system.syslog"
  lifecycle_name  = elasticstack_elasticsearch_index_lifecycle.test.name
}
