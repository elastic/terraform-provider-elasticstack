variable "name" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
  fleet {}
}

resource "elasticstack_fleet_agent_policy" "apl-http-monitor-labels" {
  name            = "TestMonitorResource Agent Policy - ${var.name}"
  namespace       = "testacc"
  description     = "TestMonitorResource Agent Policy"
  monitor_logs    = true
  monitor_metrics = true
  skip_destroy    = false
}

resource "elasticstack_kibana_synthetics_private_location" "pl-http-monitor-labels" {
  label           = "monitor-pll-${var.name}"
  agent_policy_id = elasticstack_fleet_agent_policy.apl-http-monitor-labels.policy_id
}

resource "elasticstack_kibana_synthetics_monitor" "http-monitor-labels" {
  name              = "TestHttpMonitorLabels Removed - ${var.name}"
  private_locations = [elasticstack_kibana_synthetics_private_location.pl-http-monitor-labels.label]
  http = {
    url = "http://localhost:5601"
  }
}
