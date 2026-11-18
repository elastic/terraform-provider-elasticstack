variable "description" {
  type = string
}

variable "policy_name" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_output" "test_output" {
  name                 = "${var.policy_name} Output"
  output_id            = "${var.policy_name}-output"
  type                 = "elasticsearch"
  default_integrations = false
  default_monitoring   = false
  hosts                = ["https://elasticsearch:9200"]
}

# data_output_id and monitoring_output_id are intentionally not configured: the
# test assigns the output outside Terraform. depends_on makes sure the policy is
# destroyed before the output, which Fleet refuses to delete while a policy
# still uses it.
resource "elasticstack_fleet_agent_policy" "test_policy" {
  name            = "Policy ${var.policy_name}"
  namespace       = "default"
  description     = var.description
  monitor_logs    = false
  monitor_metrics = false

  depends_on = [elasticstack_fleet_output.test_output]
}
