variable "name" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
  fleet {}
}

resource "elasticstack_fleet_agent_policy" "apl-http-monitor-ssl" {
  name               = "TestMonitorResource Agent Policy - ${var.name}"
  namespace          = "testacc"
  description        = "TestMonitorResource Agent Policy"
  monitor_logs       = true
  monitor_metrics    = true
  skip_destroy       = false
  download_source_id = elasticstack_fleet_agent_download_source.default.source_id
}

resource "elasticstack_fleet_agent_download_source" "default" {
  name      = "Agent Download Source HTTP Monitor ${var.name}"
  source_id = "agent-download-source-http-monitor-${var.name}"
  default   = false
  host      = "https://artifacts.elastic.co/downloads/elastic-agent"
  space_ids = ["default"]
}

resource "elasticstack_kibana_synthetics_private_location" "pl-http-monitor-ssl" {
  label           = "monitor-pll-${var.name}"
  agent_policy_id = elasticstack_fleet_agent_policy.apl-http-monitor-ssl.policy_id
}

resource "elasticstack_kibana_synthetics_monitor" "http-monitor-ssl" {
  name              = "TestHttpMonitorResource - ${var.name}"
  private_locations = [elasticstack_kibana_synthetics_private_location.pl-http-monitor-ssl.label]
  http = {
    url                         = "http://localhost:5601"
    ssl_verification_mode       = "full"
    ssl_supported_protocols     = ["TLSv1.2"]
    ssl_certificate_authorities = ["ca1", "ca2"]
    ssl_certificate             = "cert"
    ssl_key                     = "key"
    ssl_key_passphrase          = "pass"
  }
}
