provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_output" "test_output" {
  name                 = "Output ${var.policy_name}"
  output_id            = "${var.policy_name}-output"
  type                 = "elasticsearch"
  hosts                = ["https://elasticsearch:9200"]
  default_integrations = false
  default_monitoring   = false
}

resource "elasticstack_fleet_agent_policy" "test_policy" {
  name                 = "Policy ${var.policy_name}"
  namespace            = "default"
  description          = "Test Agent Policy with Output IDs"
  monitor_logs         = false
  monitor_metrics      = false
  data_output_id       = elasticstack_fleet_output.test_output.output_id
  monitoring_output_id = elasticstack_fleet_output.test_output.output_id
}
