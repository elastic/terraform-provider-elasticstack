provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_output" "test" {
  name                 = "test-logstash"
  type                 = "logstash"
  hosts                = ["logstash:5044", "logstash2:5044"]
  default_integrations = false
  default_monitoring   = false
  ssl = {
    certificate_authorities = ["placeholder"]
    certificate             = "placeholder"
    key                     = "placeholder"
  }
}

data "elasticstack_fleet_output" "test" {}
