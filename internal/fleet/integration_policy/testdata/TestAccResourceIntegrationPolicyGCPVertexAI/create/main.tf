variable "policy_name" {
  description = "The integration policy name"
  type        = string
}

resource "elasticstack_fleet_agent_policy" "test" {
  name            = "${var.policy_name} Agent policy"
  namespace       = "default"
  description     = "Test Agent Policy Vertex AI"
  monitor_logs    = true
  monitor_metrics = true
}

resource "elasticstack_fleet_integration_policy" "test_policy" {
  name                = var.policy_name
  namespace           = "default"
  description         = "IntegrationPolicyTest Policy Vertex AI"
  integration_name    = "gcp_vertexai"
  integration_version = "1.4.0"
  agent_policy_id     = elasticstack_fleet_agent_policy.test.id

  vars_json = jsonencode({
    project_id = "my-gcp-project"
  })

  inputs = {
    "GCP Vertex AI  Logs-gcp/metrics" = {
      enabled = true,
      streams = {
        "gcp_vertexai.prompt_response_logs" = {
          enabled = true,
          vars = jsonencode({
            period              = "300s",
            table_id            = "table_id",
            time_lookback_hours = 1,
            exclude_labels      = false,
            tags = [
              "forwarded",
              "gcp-vertexai-prompt-response-logs"
            ]
          })
        }
      }
    },
    "GCP Vertex AI Metrics-gcp/metrics" = {
      enabled = true
      streams = {
        "gcp_vertexai.metrics" = {
          enabled = true
          vars = jsonencode({
            period  = "60s"
            regions = ["us-central1"]
          })
        }
      }
    }
  }
}
