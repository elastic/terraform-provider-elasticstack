variable "policy_name" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_integration" "test_policy" {
  name    = "system"
  version = "1.64.0"
  force   = true
}

resource "elasticstack_fleet_agent_policy" "test_policy" {
  name            = "${var.policy_name} Agent Policy"
  namespace       = "default"
  monitor_logs    = false
  monitor_metrics = false
  skip_destroy    = false
}

resource "elasticstack_fleet_output" "test_output" {
  name                 = "${var.policy_name} Output"
  output_id            = "${var.policy_name}-output"
  type                 = "elasticsearch"
  default_integrations = false
  default_monitoring   = false
  hosts                = ["https://elasticsearch:9200"]
}

# output_id has been removed from configuration. depends_on makes sure the
# policy is destroyed before the output, which Fleet refuses to delete while a
# policy still uses it.
resource "elasticstack_fleet_integration_policy" "test_policy" {
  name                = var.policy_name
  namespace           = "default"
  agent_policy_id     = elasticstack_fleet_agent_policy.test_policy.policy_id
  integration_name    = elasticstack_fleet_integration.test_policy.name
  integration_version = elasticstack_fleet_integration.test_policy.version

  depends_on = [elasticstack_fleet_output.test_output]
}
