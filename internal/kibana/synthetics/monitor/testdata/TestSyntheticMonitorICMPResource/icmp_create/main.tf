variable "name" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
  fleet {}
}

resource "elasticstack_fleet_agent_policy" "apl-icmp-monitor" {
  name            = "TestMonitorResource Agent Policy - ${var.name}"
  namespace       = "testacc"
  description     = "TestMonitorResource Agent Policy"
  monitor_logs    = true
  monitor_metrics = true
  skip_destroy    = false
}

resource "elasticstack_kibana_synthetics_private_location" "pl-icmp-monitor" {
  label           = "monitor-pll-${var.name}"
  agent_policy_id = elasticstack_fleet_agent_policy.apl-icmp-monitor.policy_id
}

resource "elasticstack_kibana_synthetics_monitor" "icmp-monitor" {
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
