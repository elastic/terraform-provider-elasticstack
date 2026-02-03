provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

# Install a Fleet integration - this creates Fleet-managed index templates
resource "elasticstack_fleet_integration" "tcp" {
  name         = "tcp"
  version      = "1.16.0"
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

# Attach the ILM policy to the Fleet-managed template
# The TCP integration creates the "logs-tcp.generic" index template
resource "elasticstack_elasticsearch_index_template_ilm_attachment" "test" {
  depends_on = [elasticstack_fleet_integration.tcp]

  index_template = "logs-tcp.generic"
  lifecycle_name = elasticstack_elasticsearch_index_lifecycle.test.name
}
