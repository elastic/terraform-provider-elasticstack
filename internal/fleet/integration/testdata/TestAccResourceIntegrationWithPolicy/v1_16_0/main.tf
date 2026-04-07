variable "policy_name" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_integration" "test_integration" {
  name         = "tcp"
  version      = "1.16.0"
  force        = true
  skip_destroy = true
}

# An agent policy to hold the integration policy.
resource "elasticstack_fleet_agent_policy" "sample" {
  name            = var.policy_name
  namespace       = "default"
  description     = "A sample agent policy"
  monitor_logs    = true
  monitor_metrics = true
  skip_destroy    = false
}

# The associated enrollment token.
data "elasticstack_fleet_enrollment_tokens" "sample" {
  policy_id = elasticstack_fleet_agent_policy.sample.policy_id
}

# The integration policy.
resource "elasticstack_fleet_integration_policy" "sample" {
  name                = var.policy_name
  namespace           = "default"
  description         = "A sample integration policy"
  agent_policy_id     = elasticstack_fleet_agent_policy.sample.policy_id
  integration_name    = elasticstack_fleet_integration.test_integration.name
  integration_version = elasticstack_fleet_integration.test_integration.version

  inputs = {
    "tcp-tcp" = {
      streams = {
        "tcp.generic" = {
          enabled = true,
          vars = jsonencode({
            "listen_address" : "localhost",
            "listen_port" : 8080,
            "data_stream.dataset" : "tcp.generic",
            "tags" : [],
            "syslog_options" : "field: message\n#format: auto\n#timezone: Local\n",
            "ssl" : "#certificate: |\n#    -----BEGIN CERTIFICATE-----\n#    ...\n#    -----END CERTIFICATE-----\n#key: |\n#    -----BEGIN PRIVATE KEY-----\n#    ...\n#    -----END PRIVATE KEY-----\n",
            "custom" : ""
          })
        }
      }
    }
  }
}
