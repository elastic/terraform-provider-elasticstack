variable "policy_name" {
  type = string
}

resource "elasticstack_fleet_agent_policy" "test_policy" {
  name            = var.policy_name
  namespace       = "default"
  description     = "TestAccIntegrationPolicyInputs Agent Policy"
  monitor_logs    = true
  monitor_metrics = true
  skip_destroy    = false
}

data "elasticstack_fleet_integration" "test" {
  name = "azure_metrics"
}

resource "elasticstack_fleet_integration_policy" "test_policy" {
  name                = var.policy_name
  namespace           = "default"
  agent_policy_id     = elasticstack_fleet_agent_policy.test_policy.id
  integration_name    = "azure_metrics"
  integration_version = data.elasticstack_fleet_integration.test.version
  description         = "Azure Metrics Integration Policy"
  vars_json = jsonencode({
    client_id       = "test-client-id",
    tenant_id       = "test-tenant-id",
    client_secret   = "test-client-secret",
    subscription_id = "test-subscription-id",
  })

  inputs = {
    "monitor-azure/metrics" = {
      enabled = true,
      streams = {
        "azure.monitor" = {
          enabled = true,
          vars = jsonencode({
            period    = "300s",
            resources = "- resource_query: \"resourceType eq 'Microsoft.Search/searchServices'\"\n  metrics:\n  - name: [\"DocumentsProcessedCount\", \"SearchLatency\", \"SearchQueriesPerSecond\", \"ThrottledSearchQueriesPercentage\"]\n    namespace: \"Microsoft.Search/searchServices\""
          })
        }
      }
    }
  }
}
