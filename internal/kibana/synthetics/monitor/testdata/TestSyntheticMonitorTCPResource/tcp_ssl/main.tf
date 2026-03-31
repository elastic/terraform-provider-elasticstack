variable "name" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
  fleet {}
}

resource "elasticstack_fleet_agent_policy" "apl-tcp-monitor-ssl" {
  name            = "TestMonitorResource Agent Policy - ${var.name}"
  namespace       = "testacc"
  description     = "TestMonitorResource Agent Policy"
  monitor_logs    = true
  monitor_metrics = true
  skip_destroy    = false
}

resource "elasticstack_kibana_synthetics_private_location" "pl-tcp-monitor-ssl" {
  label           = "monitor-pll-${var.name}"
  agent_policy_id = elasticstack_fleet_agent_policy.apl-tcp-monitor-ssl.policy_id
}

resource "elasticstack_kibana_synthetics_monitor" "tcp-monitor-ssl" {
  name              = "TestHttpMonitorResource - ${var.name}"
  private_locations = [elasticstack_kibana_synthetics_private_location.pl-tcp-monitor-ssl.label]
  tcp = {
    host                         = "http://localhost:5601"
    ssl_verification_mode        = "full"
    ssl_supported_protocols      = ["TLSv1.2"]
    ssl_certificate_authorities  = ["ca1", "ca2"]
    ssl_certificate              = "cert"
    ssl_key                      = "key"
    ssl_key_passphrase           = "pass"
  }
}
