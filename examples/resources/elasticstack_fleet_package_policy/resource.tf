provider "elasticstack" {
  fleet {}
}

// The package to use.
resource "elasticstack_fleet_package" "sample" {
  name    = "tcp"
  version = "1.16.0"
  force   = true
}

// An agent policy to hold the package policy.
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

// The package policy.
resource "elasticstack_fleet_package_policy" "sample" {
  name            = "Sample Package Policy"
  namespace       = "default"
  description     = "A sample package policy"
  agent_policy_id = elasticstack_fleet_agent_policy.sample.policy_id
  package_name    = elasticstack_fleet_package.sample.name
  package_version = elasticstack_fleet_package.sample.version

  input {
    input_id = "tcp-tcp"
    streams_json = jsonencode({
      "tcp.generic" : {
        "enabled" : true,
        "vars" : {
          "listen_address" : "localhost",
          "listen_port" : 8080,
          "data_stream.dataset" : "tcp.generic",
          "tags" : [],
          "syslog_options" : "field: message\n#format: auto\n#timezone: Local\n",
          "ssl" : "#certificate: |\n#    -----BEGIN CERTIFICATE-----\n#    ...\n#    -----END CERTIFICATE-----\n#key: |\n#    -----BEGIN PRIVATE KEY-----\n#    ...\n#    -----END PRIVATE KEY-----\n",
          "custom" : ""
        }
      }
    })
  }
}
