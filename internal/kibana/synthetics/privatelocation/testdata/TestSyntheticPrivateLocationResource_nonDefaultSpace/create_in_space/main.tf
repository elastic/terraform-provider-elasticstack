variable "suffix" {
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
  name        = "acc-synthetics-pl-${var.space_id}"
  description = "Kibana space for synthetics private location acceptance test"
}

resource "elasticstack_fleet_agent_download_source" "default" {
  name      = "Agent Download Source Private Location Space ${var.suffix}"
  source_id = "agent-download-source-private-location-space-${var.suffix}"
  default   = false
  host      = "https://artifacts.elastic.co/downloads/elastic-agent"
  space_ids = ["default"]
}

resource "elasticstack_fleet_agent_policy" "test_policy" {
  name               = "Private Location Agent Policy - test_policy - ${var.suffix}"
  namespace          = replace(var.space_id, "-", "_")
  description        = "TestPrivateLocationResource Agent Policy"
  monitor_logs       = true
  monitor_metrics    = true
  skip_destroy       = false
  space_ids          = [elasticstack_kibana_space.test.space_id]
  download_source_id = elasticstack_fleet_agent_download_source.default.source_id
}

resource "elasticstack_kibana_synthetics_private_location" "test" {
  space_id        = elasticstack_kibana_space.test.space_id
  label           = "pl-test-label-space-${var.suffix}"
  agent_policy_id = elasticstack_fleet_agent_policy.test_policy.policy_id
  tags            = ["a", "b"]
  geo = {
    lat = 42.42
    lon = -42.42
  }
}
