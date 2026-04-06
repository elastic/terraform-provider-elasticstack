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
  name        = "acc-synthetics-tcp-${var.space_id}"
  description = "Kibana space for synthetics TCP monitor acceptance test"
}

resource "elasticstack_fleet_agent_policy" "apl-tcp-monitor" {
  name            = "TestMonitorResource Agent Policy - ${var.name}"
  namespace       = replace(var.space_id, "-", "_")
  description     = "TestMonitorResource Agent Policy"
  monitor_logs    = true
  monitor_metrics = true
  skip_destroy    = false
  space_ids       = [elasticstack_kibana_space.test.space_id]
}

resource "elasticstack_kibana_synthetics_private_location" "pl-tcp-monitor" {
  space_id        = elasticstack_kibana_space.test.space_id
  label           = "monitor-pll-ns-${var.name}"
  agent_policy_id = elasticstack_fleet_agent_policy.apl-tcp-monitor.policy_id
}

resource "elasticstack_kibana_synthetics_monitor" "tcp-monitor" {
  space_id          = elasticstack_kibana_space.test.space_id
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
