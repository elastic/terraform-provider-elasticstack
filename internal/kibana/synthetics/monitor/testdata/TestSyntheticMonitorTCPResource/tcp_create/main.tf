variable "name" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
  fleet {}
}

resource "elasticstack_fleet_agent_policy" "apl-tcp-monitor" {
  name               = "TestMonitorResource Agent Policy - ${var.name}"
  namespace          = "testacc"
  description        = "TestMonitorResource Agent Policy"
  monitor_logs       = true
  monitor_metrics    = true
  skip_destroy       = false
  download_source_id = elasticstack_fleet_agent_download_source.default.source_id
}

resource "elasticstack_fleet_agent_download_source" "default" {
  name      = "Agent Download Source TCP Monitor ${var.name}"
  source_id = "agent-download-source-tcp-monitor-${var.name}"
  default   = false
  host      = "https://artifacts.elastic.co/downloads/elastic-agent"
  space_ids = ["default"]
}

resource "elasticstack_kibana_synthetics_private_location" "pl-tcp-monitor" {
  label           = "monitor-pll-${var.name}"
  agent_policy_id = elasticstack_fleet_agent_policy.apl-tcp-monitor.policy_id
}

resource "elasticstack_kibana_synthetics_monitor" "tcp-monitor" {
  name              = "TestTcpMonitorResource - ${var.name}"
  namespace         = "testacc_test"
  schedule          = 5
  private_locations = [elasticstack_kibana_synthetics_private_location.pl-tcp-monitor.label]
  enabled           = true
  tags              = ["a", "b"]
  alert = {
    status = {
      enabled = true
    }
    tls = {
      enabled = true
    }
  }
  service_name = "test apm service"
  timeout      = 30
  tcp = {
    host                     = "http://localhost:5601"
    proxy_use_local_resolver = true
  }
}
