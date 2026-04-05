provider "elasticstack" {
  kibana {}
}

resource "elasticstack_fleet_output" "logstash_output" {
  name  = "Logstash Output"
  type  = "logstash"
  hosts = ["logstash:5044"]

  default_integrations = false
  default_monitoring   = false

  ssl = {
    certificate_authorities = ["placeholder"]
    certificate             = "placeholder"
    key                     = "placeholder"
  }

  config_yaml = yamlencode({
    "ssl.verification_mode" = "none"
  })
}
