variable "name" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
  fleet {}
}

resource "elasticstack_fleet_agent_download_source" "default" {
  name      = "Agent Download Source Monitor Labels ${var.name}"
  source_id = "agent-download-source-monitor-labels-${var.name}"
  default   = true
  host      = "https://artifacts.elastic.co/downloads/elastic-agent"
  space_ids = ["default"]
}

resource "elasticstack_fleet_agent_policy" "apl-http-monitor-labels" {
  name               = "TestMonitorResource Agent Policy - ${var.name}"
  namespace          = "testacc"
  description        = "TestMonitorResource Agent Policy"
  monitor_logs       = true
  monitor_metrics    = true
  skip_destroy       = false
  download_source_id = elasticstack_fleet_agent_download_source.default.source_id
}

resource "elasticstack_kibana_synthetics_private_location" "pl-http-monitor-labels" {
  label           = "monitor-pll-${var.name}"
  agent_policy_id = elasticstack_fleet_agent_policy.apl-http-monitor-labels.policy_id
}

resource "elasticstack_kibana_synthetics_monitor" "http-monitor-labels" {
  name              = "TestHttpMonitorLabels - ${var.name}"
  private_locations = [elasticstack_kibana_synthetics_private_location.pl-http-monitor-labels.label]
  labels = {
    environment = "production"
    team        = "platform"
    service     = "web-app"
  }
  http = {
    url = "http://localhost:5601"
  }
}
