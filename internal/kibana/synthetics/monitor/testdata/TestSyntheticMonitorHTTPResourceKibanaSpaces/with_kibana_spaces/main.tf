variable "name" {
  type = string
}

variable "visibility_space_id" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
  fleet {}
}

resource "elasticstack_kibana_space" "visibility" {
  space_id    = var.visibility_space_id
  name        = "acc-synthetics-visibility-${var.visibility_space_id}"
  description = "Kibana space for Synthetics monitor visibility acceptance test"
}

resource "elasticstack_fleet_agent_policy" "apl-http-monitor" {
	name               = "TestMonitorResource Agent Policy - ${var.name}"
	namespace          = "testacc"
	description        = "TestMonitorResource Agent Policy"
	monitor_logs       = true
	monitor_metrics    = true
	skip_destroy       = false
	space_ids          = ["default", elasticstack_kibana_space.visibility.space_id]
	download_source_id = elasticstack_fleet_agent_download_source.default.source_id
}

resource "elasticstack_fleet_agent_download_source" "default" {
  name      = "Agent Download Source HTTP Monitor ${var.name}"
  source_id = "agent-download-source-http-monitor-${var.name}"
	default   = false
	host      = "https://artifacts.elastic.co/downloads/elastic-agent"
	space_ids = ["default", elasticstack_kibana_space.visibility.space_id]
}

resource "elasticstack_kibana_synthetics_private_location" "monitor" {
  label           = "monitor-pll-${var.name}"
  agent_policy_id = elasticstack_fleet_agent_policy.apl-http-monitor.policy_id
}

resource "elasticstack_kibana_synthetics_monitor" "http-monitor" {
  name              = "TestHttpMonitorResource - ${var.name}"
  private_locations = [elasticstack_kibana_synthetics_private_location.monitor.label]
  kibana_spaces     = [elasticstack_kibana_space.visibility.space_id]

  http = {
    url = "http://localhost:5601"
  }
}
