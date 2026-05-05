provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_output" "test_output" {
  name                 = "Test Output ${var.policy_name}"
  output_id            = "${var.policy_name}-output"
  type                 = "elasticsearch"
  default_integrations = false
  default_monitoring   = false
  hosts = [
    "https://elasticsearch:9200"
  ]
}

resource "elasticstack_fleet_agent_policy" "test_policy" {
  name                = var.policy_name
  namespace           = "default"
  description         = "Test Agent Policy with Output IDs"
  monitor_logs        = true
  monitor_metrics     = false
  data_output_id      = elasticstack_fleet_output.test_output.output_id
  monitoring_output_id = elasticstack_fleet_output.test_output.output_id
}
