variable "name" {
  type = string
}

variable "space_id" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
  fleet {}
}

resource "elasticstack_kibana_space" "test" {
  space_id    = var.space_id
  name        = "acc-synthetics-icmp-${var.space_id}"
  description = "Kibana space for synthetics ICMP monitor acceptance test"
}

resource "elasticstack_fleet_agent_policy" "apl-icmp-monitor" {
  name               = "TestMonitorResource Agent Policy - ${var.name}"
  namespace          = replace(var.space_id, "-", "_")
  description        = "TestMonitorResource Agent Policy"
  monitor_logs       = true
  monitor_metrics    = true
  skip_destroy       = false
  space_ids          = [elasticstack_kibana_space.test.space_id]
  download_source_id = elasticstack_fleet_agent_download_source.default.source_id
}

resource "elasticstack_fleet_agent_download_source" "default" {
  name      = "Agent Download Source ICMP Monitor ${var.name}"
  source_id = "agent-download-source-icmp-monitor-${var.name}"
  default   = false
  host      = "https://artifacts.elastic.co/downloads/elastic-agent"
  space_ids = [elasticstack_kibana_space.test.space_id]
}

resource "elasticstack_kibana_synthetics_private_location" "pl-icmp-monitor" {
  space_id        = elasticstack_kibana_space.test.space_id
  label           = "monitor-pll-ns-${var.name}"
  agent_policy_id = elasticstack_fleet_agent_policy.apl-icmp-monitor.policy_id
}

resource "elasticstack_kibana_synthetics_monitor" "icmp-monitor" {
  space_id          = elasticstack_kibana_space.test.space_id
  name              = "TestIcmpMonitorResource - ${var.name}"
  namespace         = "testacc_namespace"
  schedule          = 5
  private_locations = [elasticstack_kibana_synthetics_private_location.pl-icmp-monitor.label]
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
  icmp = {
    host = "localhost"
  }
}
