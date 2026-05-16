provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_integration" "test_integration" {
  name         = "system"
  version      = "1.18.0"
  force        = true
  skip_destroy = true
}

resource "elasticstack_elasticsearch_index_lifecycle" "test" {
  name          = "test-fleet-ilm-policy"
  force_destroy = true

  hot {
    rollover {
      max_age = "1d"
    }
  }
}

resource "elasticstack_elasticsearch_index_template_ilm_attachment" "test" {
  depends_on = [elasticstack_fleet_integration.test_integration]

  index_template = "logs-system.syslog"
  lifecycle_name = elasticstack_elasticsearch_index_lifecycle.test.name
}
