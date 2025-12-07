variable "policy_name" {
  type        = string
  description = "Name for the agent policy"
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_agent_policy" "test_policy" {
  name            = var.policy_name
  namespace       = "default"
  description     = "Test Agent Policy with Multiple Required Versions"
  monitor_logs    = true
  monitor_metrics = false
  skip_destroy    = false
  required_versions = {
    "8.15.0" = 50
    "8.16.0" = 50
  }
}
