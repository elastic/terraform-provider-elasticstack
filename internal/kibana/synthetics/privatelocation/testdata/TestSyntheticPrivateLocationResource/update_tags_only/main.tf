variable "suffix" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
  fleet {}
}

resource "elasticstack_fleet_agent_download_source" "default" {
  name      = "Agent Download Source Private Location ${var.suffix}"
  source_id = "agent-download-source-private-location-${var.suffix}"
  default   = false
  host      = "https://artifacts.elastic.co/downloads/elastic-agent"
  space_ids = ["default"]
}

resource "elasticstack_fleet_agent_policy" "test_policy_default" {
  name               = "Private Location Agent Policy - test_policy_default - ${var.suffix}"
  namespace          = "default"
  description        = "TestPrivateLocationResource Agent Policy"
  monitor_logs       = true
  monitor_metrics    = true
  skip_destroy       = false
  download_source_id = elasticstack_fleet_agent_download_source.default.source_id
}

resource "elasticstack_kibana_synthetics_private_location" "test" {
  label           = "pl-test-label-2-${var.suffix}"
  agent_policy_id = elasticstack_fleet_agent_policy.test_policy_default.policy_id
  tags            = ["c", "d", "e"]
}
