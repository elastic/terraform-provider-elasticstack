provider "elasticstack" {
  fleet {}
}

// The integration to use.
resource "elasticstack_fleet_integration" "sample" {
  name    = "tcp"
  version = "1.16.0"
  force   = true
}

// An agent policy to hold the integration policy.
resource "elasticstack_fleet_agent_policy" "sample" {
  name            = "Sample Agent Policy"
  namespace       = "default"
  description     = "A sample agent policy"
  monitor_logs    = true
  monitor_metrics = true
  skip_destroy    = false
}

// The associated enrollment token.
data "elasticstack_fleet_enrollment_tokens" "sample" {
  policy_id = elasticstack_fleet_agent_policy.sample.policy_id
}

// The integration policy.
resource "elasticstack_fleet_integration_policy" "sample" {
  name                = "Sample Integration Policy"
  namespace           = "default"
  description         = "A sample integration policy"
  agent_policy_id     = elasticstack_fleet_agent_policy.sample.policy_id
  integration_name    = elasticstack_fleet_integration.sample.name
  integration_version = elasticstack_fleet_integration.sample.version
  // Optional: specify a custom output to send data to
  // output_id           = "my-custom-output-id"

  inputs = {
    "tcp-tcp" = {
      enabled = true
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
